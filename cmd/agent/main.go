package main

import (
	"context"
	"flag"
	"time"

	"github.com/caarlos0/env/v7"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/custom"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/gopsutil"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/provider/runtime"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/pusher/http"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
)

var (
	buildVersion                 = "N/A"
	buildDate                    = "N/A"
	buildCommit                  = "N/A"
	defaultPushRateLimit         = 20
	defaultPushTimeout           = 10 * time.Second
	defaultSendMetricsInterval   = 10 * time.Second
	defaultUpdateMetricsInterval = 2 * time.Second
)

type config struct {
	Key                   string `env:"KEY"`
	ServerURL             string `env:"ADDRESS"`
	CollectMetricsList    []string
	PushRateLimit         int           `env:"RATE_LIMIT"`
	PushTimeout           time.Duration `env:"PUSH_TIMEOUT"`
	SendMetricsInterval   time.Duration `env:"REPORT_INTERVAL"`
	UpdateMetricsInterval time.Duration `env:"POLL_INTERVAL"`
}

func main() {
	logger.InfoFormat("Build version: %s\n", buildVersion)
	logger.InfoFormat("Build date: %s\n", buildDate)
	logger.InfoFormat("Build commit: %s\n", buildCommit)

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("initialize config", err))
	}

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	metricPusher, err := http.NewMetricsPusher(conf, converter, nil)
	if err != nil {
		panic(logger.WrapError("create new metrics pusher", err))
	}

	runtimeMetricsProvider := runtime.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := custom.NewCustomMetricsProvider()
	gopsutilMetricsProvider := gopsutil.NewGopsutilMetricsProvider()
	aggregateMetricsProvider := provider.NewAggregateMetricsProvider(
		runtimeMetricsProvider,
		customMetricsProvider,
		gopsutilMetricsProvider,
	)
	getMetricsWorker := worker.NewPeriodicWorker(aggregateMetricsProvider.Update)
	pushMetricsWorker := worker.NewPeriodicWorker(func(workerContext context.Context) error {
		return metricPusher.Push(workerContext, aggregateMetricsProvider.GetMetrics())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go getMetricsWorker.StartWork(ctx, conf.UpdateMetricsInterval)
	pushMetricsWorker.StartWork(ctx, conf.SendMetricsInterval)
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

	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Metrics server URL")
	flag.IntVar(&conf.PushRateLimit, "l", defaultPushRateLimit, "Push metrics parallel workers limit")
	flag.DurationVar(&conf.PushTimeout, "t", defaultPushTimeout, "Push metrics timeout")
	flag.DurationVar(&conf.SendMetricsInterval, "r", defaultSendMetricsInterval, "Send metrics interval")
	flag.DurationVar(&conf.UpdateMetricsInterval, "p", defaultUpdateMetricsInterval, "Update metrics interval")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
}

func (c *config) MetricsList() []string {
	return c.CollectMetricsList
}

func (c *config) MetricsServerURL() string {
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
