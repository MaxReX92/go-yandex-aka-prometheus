package main

import (
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"time"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto/rsa"
	"github.com/caarlos0/env/v7"

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
)

var (
	buildVersion         = "N/A"
	buildDate            = "N/A"
	buildCommit          = "N/A"
	defaultStoreInterval = 300 * time.Second
)

type config struct {
	CryptoKey     string        `env:"CRYPTO_KEY"`
	Key           string        `env:"KEY"`
	ServerURL     string        `env:"ADDRESS"`
	StoreFile     string        `env:"STORE_FILE"`
	DB            string        `env:"DATABASE_DSN"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Restore       bool          `env:"RESTORE"`
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

	if conf.Restore {
		logger.Info("Restore metrics from backup")
		err = storageStrategy.RestoreFromBackup(ctx)
		if err != nil {
			logger.ErrorFormat("failed to restore state from backup: %v", err)
		}
	}

	if !conf.SyncMode() {
		logger.Info("Start periodic backup serice")
		backgroundStore := worker.NewPeriodicWorker(func(ctx context.Context) error { return storageStrategy.CreateBackup(ctx) })
		go backgroundStore.StartWork(ctx, conf.StoreInterval)
	}

	err = metricsServer.Start()
	if err != nil {
		logger.ErrorObj(err)
	}
}

func createConfig() (*config, error) {
	conf := &config{}

	flag.StringVar(&conf.CryptoKey, "crypto-key", "", "Server private crypto key path")
	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.BoolVar(&conf.Restore, "r", true, "Restore metric values from the server backup file")
	flag.DurationVar(&conf.StoreInterval, "i", defaultStoreInterval, "Store backup interval")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Server listen URL")
	flag.StringVar(&conf.StoreFile, "f", "/tmp/devops-metrics-dataBase.json", "Backup storage file path")
	flag.StringVar(&conf.DB, "d", "", "Database connection stirng")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
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
