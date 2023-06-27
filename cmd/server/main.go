package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v7"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto/rsa"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database/postgres"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database/stub"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/server"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/db"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/file"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage/memory"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
	"github.com/MaxReX92/go-yandex-aka-prometheus/pkg/runner"
)

var (
	buildVersion         = "N/A"
	buildDate            = "N/A"
	buildCommit          = "N/A"
	defaultStoreInterval = 300 * time.Second
)

type config struct {
	ConfigPath    string        `env:"CONFIG"`
	CryptoKey     string        `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	Key           string        `env:"KEY" json:"key,omitempty"`
	ServerURL     string        `env:"ADDRESS" json:"address,omitempty"`
	StoreFile     string        `env:"STORE_FILE" json:"store_file,omitempty"`
	DB            string        `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	Restore       bool          `env:"RESTORE" json:"restore,omitempty"`
}

func main() {
	logger.InfoFormat("Build version: %s\n", buildVersion)
	logger.InfoFormat("Build date: %s\n", buildDate)
	logger.InfoFormat("Build commit: %s\n", buildCommit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("create config file", err))
	}
	logger.InfoFormat("Starting server with the following configuration:%v", conf)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var base database.DataBase
	var backupStorage storage.MetricsStorage
	if conf.DB == "" {
		base = &stub.StubDataBase{}
		backupStorage = file.NewFileStorage(conf)
	} else {
		base, err = postgres.NewPostgresDataBase(ctx, conf)
		if err != nil {
			panic(logger.WrapError("create database", err))
		}

		backupStorage = db.NewDBStorage(base)
	}
	defer base.Close()

	inMemoryStorage := memory.NewInMemoryStorage()
	storageStrategy := storage.NewStorageStrategy(conf, inMemoryStorage, backupStorage)
	defer storageStrategy.Close()

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	htmlPageBuilder := html.NewSimplePageBuilder()

	var decryptor crypto.Decryptor
	if conf.CryptoKey != "" {
		decryptor, err = rsa.NewDecryptor(conf.CryptoKey)
		if err != nil {
			panic(logger.WrapError("create decryptor", err))
		}
	}

	metricsServer := server.New(conf, storageStrategy, converter, htmlPageBuilder, base, decryptor)
	runners := []runner.Runner{metricsServer}

	if conf.Restore {
		logger.Info("Restore metrics from backup")
		err = storageStrategy.RestoreFromBackup(ctx)
		if err != nil {
			logger.ErrorFormat("failed to restore state from backup: %v", err)
		}
	}

	if !conf.SyncMode() {
		logger.Info("Start periodic backup serice")
		backgroundStore := worker.NewPeriodicWorker(conf.StoreInterval, func(ctx context.Context) error { return storageStrategy.CreateBackup(ctx) })
		runners = append(runners, &backgroundStore)
	}

	multiRunner := runner.NewMultiWorker(runners...)
	gracefulRunner := runner.NewGracefulRunner(multiRunner)
	gracefulRunner.Start(ctx)

	// shutdown
	select {
	case err = <-gracefulRunner.Error():
		err = logger.WrapError("start application", err)
	case <-interrupt:
		err = gracefulRunner.Stop(ctx)
	}

	if err != nil {
		logger.ErrorObj(err)
	}
}

func createConfig() (*config, error) {
	conf := &config{}

	flag.StringVar(&conf.ConfigPath, "c", "", "Json config file path")
	flag.StringVar(&conf.ConfigPath, "config", "", "Json config file path")
	flag.StringVar(&conf.CryptoKey, "crypto-key", "", "Server private crypto key path")
	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.BoolVar(&conf.Restore, "r", true, "Restore metric values from the server backup file")
	flag.DurationVar(&conf.StoreInterval, "i", defaultStoreInterval, "Store backup interval")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Server listen URL")
	flag.StringVar(&conf.StoreFile, "f", "/tmp/devops-metrics-dataBase.json", "Backup storage file path")
	flag.StringVar(&conf.DB, "d", "", "Database connection stirng")
	flag.Parse()

	err := env.Parse(conf)
	if err != nil {
		return nil, logger.WrapError("parse flags", err)
	}

	if conf.ConfigPath != "" {
		content, err := os.ReadFile(conf.ConfigPath)
		if err != nil {
			return nil, logger.WrapError("read json config file", err)
		}

		err = json.Unmarshal(content, conf)
		if err != nil {
			return nil, logger.WrapError("unmarshal json config file", err)
		}
	}

	return conf, nil
}

func (c *config) ListenURL() string {
	return c.ServerURL
}

func (c *config) StoreFilePath() string {
	return c.StoreFile
}

func (c *config) SyncMode() bool {
	return c.DB != "" || c.StoreInterval == 0
}

func (c *config) String() string {
	return fmt.Sprintf("\nServerURL:\t%v\nStoreInterval:\t%v\nStoreFile:\t%v\nRestore:\t%v\nDb:\t%v",
		c.ServerURL, c.StoreInterval, c.StoreFile, c.Restore, c.DB)
}

func (c *config) GetKey() []byte {
	return []byte(c.Key)
}

func (c *config) SignMetrics() bool {
	return c.Key != ""
}

func (c *config) GetConnectionString() string {
	return c.DB
}
