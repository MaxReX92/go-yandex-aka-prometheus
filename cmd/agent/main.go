package main

import (
	"context"
	"flag"
	"time"

	"github.com/caarlos0/env/v7"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/client"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
)

type config struct {
	ServerURL             string        `env:"ADDRESS"`
	PushTimeout           time.Duration `env:"PUSH_TIMEOUT"`
	SendMetricsInterval   time.Duration `env:"REPORT_INTERVAL"`
	UpdateMetricsInterval time.Duration `env:"POLL_INTERVAL"`
	CollectMetricsList    []string
}

func main() {
	conf, err := createConfig()
	if err != nil {
		panic(err)
	}
	metricPusher, err := client.NewMetricsPusher(conf)
	if err != nil {
		panic(err)
	}

	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider(runtimeMetricsProvider, customMetricsProvider)
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

	flag.StringVar(&conf.ServerURL, "a", "127.0.0.1:8080", "Metrics server URL")
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
