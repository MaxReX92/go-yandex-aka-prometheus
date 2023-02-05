package main

import (
	"context"
	"fmt"
	"go-yandex-aka-prometheus/internal/metrics"
	"go-yandex-aka-prometheus/internal/worker"
	"time"
)

func main() {
	config := metrics.RuntimeMetricsProviderConfig{
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

	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(config)
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider([]metrics.MetricsProvider{&runtimeMetricsProvider, &customMetricsProvider})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getMetricsWorker := worker.NewPeriodicWorker(worker.PeriodicWorkerConfig{Duration: 2 * time.Second}, aggregateMetricsProvider.Update)
	showMetricsWorker := worker.NewPeriodicWorker(worker.PeriodicWorkerConfig{Duration: 3 * time.Second}, func() error {
		for _, runtimeMetric := range aggregateMetricsProvider.GetMetrics() {
			fmt.Printf("%v\t\t%v\r\n", runtimeMetric.GetName(), runtimeMetric.StringValue())
		}

		// TODO: handle errors
		return nil
	})

	go getMetricsWorker.StartWork(ctx)
	showMetricsWorker.StartWork(ctx)
}
