package main

import (
	"context"
	"time"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/client"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/worker"
)

const (
	pushTimeout           = 10 * time.Second
	serverURL             = "http://127.0.0.1:8080"
	sendMetricsInterval   = 10 * time.Second
	updateMetricsInterval = 2 * time.Second
)

func main() {
	conf := createConfig()
	metricPusher := client.NewMetricsPusher(conf)
	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider([]metrics.MetricsProvider{runtimeMetricsProvider, customMetricsProvider})
	getMetricsWorker := worker.NewPeriodicWorker(aggregateMetricsProvider.Update)
	pushMetricsWorker := worker.NewPeriodicWorker(func(workerContext context.Context) error {
		return metricPusher.Push(workerContext, aggregateMetricsProvider.GetMetrics())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go getMetricsWorker.StartWork(ctx, updateMetricsInterval)
	pushMetricsWorker.StartWork(ctx, sendMetricsInterval)
}

func createConfig() *config {
	return &config{
		serverURL:             serverURL,
		pushTimeout:           pushTimeout,
		sendMetricsInterval:   sendMetricsInterval,
		updateMetricsInterval: updateMetricsInterval,
		collectMetricsList: []string{
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
		},
	}
}
