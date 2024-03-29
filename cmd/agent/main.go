package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v7"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto/rsa"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/grpc"
	grpcClient "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/grpc/client"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/http"
	httpClient "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/http/client"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/custom"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/gopsutil"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/runtime"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/pusher"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
	"github.com/MaxReX92/go-yandex-aka-prometheus/pkg/runner"
)

var (
	buildVersion                 = "N/A"
	buildDate                    = "N/A"
	buildCommit                  = "N/A"
	defaultPushRateLimit         = 20
	defaultPushTimeout           = 10 * time.Second
	defaultSendMetricsInterval   = 10 * time.Second
	defaultUpdateMetricsInterval = 2 * time.Second
	errUnkwnownChannelType       = errors.New("unknown metric channel type")
)

type config struct {
	ChannelType           string `env:"CHANNEL_TYPE" json:"channel_type,omitempty"`
	ConfigPath            string `env:"CONFIG"`
	CryptoKey             string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	Key                   string `env:"KEY" json:"key,omitempty"`
	ServerURL             string `env:"ADDRESS" json:"address,omitempty"`
	CollectMetricsList    []string
	PushRateLimit         int           `env:"RATE_LIMIT" json:"rate_limit,omitempty" `
	PushTimeout           time.Duration `env:"PUSH_TIMEOUT" json:"push_timeout,omitempty"`
	SendMetricsInterval   time.Duration `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	UpdateMetricsInterval time.Duration `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
}

func main() {
	logger.InfoFormat("Build version: %s\n", buildVersion)
	logger.InfoFormat("Build date: %s\n", buildDate)
	logger.InfoFormat("Build commit: %s\n", buildCommit)

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("initialize config", err))
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	signer := hash.NewSigner(conf)

	var encryptor crypto.Encryptor
	if conf.CryptoKey != "" {
		encryptor, err = rsa.NewEncryptor(conf.CryptoKey)
		if err != nil {
			panic(logger.WrapError("create encryptor", err))
		}
	}

	var metricPusher pusher.MetricsPusher
	switch conf.ChannelType {
	case "http":
		metricPusher, err = httpClient.NewPusher(conf, http.NewMetricsConverter(conf, signer), encryptor)
	case "grpc":
		metricPusher, err = grpcClient.NewPusher(conf, grpc.NewMetricsConverter(conf, signer))
	default:
		err = logger.WrapError(fmt.Sprintf("create new metrics pusher with type %s", conf.ChannelType), errUnkwnownChannelType)
	}
	if err != nil {
		panic(logger.WrapError("init metrics pusher", err))
	}

	runtimeMetricsProvider := runtime.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := custom.NewCustomMetricsProvider()
	gopsutilMetricsProvider := gopsutil.NewGopsutilMetricsProvider()
	aggregateMetricsProvider := provider.NewAggregateMetricsProvider(
		runtimeMetricsProvider,
		customMetricsProvider,
		gopsutilMetricsProvider,
	)
	getMetricsWorker := worker.NewPeriodicWorker(conf.UpdateMetricsInterval, aggregateMetricsProvider.Update)
	pushMetricsWorker := worker.NewPeriodicWorker(conf.SendMetricsInterval, func(workerContext context.Context) error {
		return metricPusher.Push(workerContext, aggregateMetricsProvider.GetMetrics())
	})
	multiRunner := runner.NewMultiWorker(&getMetricsWorker, &pushMetricsWorker)
	gracefulRunner := runner.NewGracefulRunner(multiRunner)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	conf := &config{CollectMetricsList: []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}}

	flag.StringVar(&conf.ChannelType, "ch", "http", "Push metrics channel type")
	flag.StringVar(&conf.ConfigPath, "c", "", "Json config file path")
	flag.StringVar(&conf.ConfigPath, "config", "", "Json config file path")
	flag.StringVar(&conf.CryptoKey, "crypto-key", "", "Agent public crypto key path")
	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Metrics server URL")
	flag.IntVar(&conf.PushRateLimit, "l", defaultPushRateLimit, "Push metrics parallel workers limit")
	flag.DurationVar(&conf.PushTimeout, "t", defaultPushTimeout, "Push metrics timeout")
	flag.DurationVar(&conf.SendMetricsInterval, "r", defaultSendMetricsInterval, "Send metrics interval")
	flag.DurationVar(&conf.UpdateMetricsInterval, "p", defaultUpdateMetricsInterval, "Update metrics interval")
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

func (c *config) MetricsList() []string {
	return c.CollectMetricsList
}

func (c *config) MetricsServerURL() string {
	return c.ServerURL
}

func (c *config) GrpcServerURL() string {
	return c.ServerURL
}

func (c *config) PushMetricsTimeout() time.Duration {
	return c.PushTimeout
}

func (c *config) ParallelLimit() int {
	return c.PushRateLimit
}

func (c *config) GetKey() []byte {
	return []byte(c.Key)
}

func (c *config) SignMetrics() bool {
	return c.Key != ""
}
