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

type config struct {
	Key                   string        `env:"KEY"`
	ServerURL             string        `env:"ADDRESS"`
	PushRateLimit         int           `env:"RATE_LIMIT"`
	PushTimeout           time.Duration `env:"PUSH_TIMEOUT"`
	SendMetricsInterval   time.Duration `env:"REPORT_INTERVAL"`
	UpdateMetricsInterval time.Duration `env:"POLL_INTERVAL"`
	CollectMetricsList    []string
}

func main() {
	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("initialize config", err))
	}

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	metricPusher, err := http.NewMetricsPusher(conf, converter)
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
	flag.IntVar(&conf.PushRateLimit, "l", 20, "Push metrics parallel workers limit")
	flag.DurationVar(&conf.PushTimeout, "t", time.Second*10, "Push metrics timeout")
	flag.DurationVar(&conf.SendMetricsInterval, "r", time.Second*10, "Send metrics interval")
	flag.DurationVar(&conf.UpdateMetricsInterval, "p", time.Second*2, "Update metrics interval")
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
