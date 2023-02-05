package main

import (
	"context"
	"go-yandex-aka-prometheus/internal"
	"go-yandex-aka-prometheus/internal/metrics"
	"go-yandex-aka-prometheus/internal/worker"
	"time"
)

const (
	updateMetricsInterval = 2 * time.Second
	sendMetricsInterval   = 10 * time.Second
)

func main() {
	metricPusher := internal.NewMetricsPusher(getMetricPusherConfig())
	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(getRuntimeMetricsConfig())
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider(
		[]metrics.MetricsProvider{&runtimeMetricsProvider, &customMetricsProvider})

	getMetricsWorker := worker.NewPeriodicWorker(
		worker.PeriodicWorkerConfig{Duration: updateMetricsInterval}, aggregateMetricsProvider.Update)
	pushMetricsWorker := worker.NewPeriodicWorker(
		worker.PeriodicWorkerConfig{Duration: sendMetricsInterval}, func(workerContext context.Context) error {
			return metricPusher.Push(workerContext, aggregateMetricsProvider.GetMetrics(workerContext))
		})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go getMetricsWorker.StartWork(ctx)
	pushMetricsWorker.StartWork(ctx)
}

func getRuntimeMetricsConfig() metrics.RuntimeMetricsProviderConfig {
	return metrics.RuntimeMetricsProviderConfig{
		MetricsList: []string{
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

func getMetricPusherConfig() internal.MetricsPusherConfig {
	return internal.MetricsPusherConfig{
		MetricsServerUrl: "127.0.0.1:8080",
	}
}
