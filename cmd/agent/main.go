package main

import (
	"fmt"
	"go-yandex-aka-prometheus/internal/metrics"
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
	err := runtimeMetricsProvider.Update()
	if err != nil {
		panic(err.Error())
	}

	for _, runtimeMetric := range runtimeMetricsProvider.GetMetrics() {
		fmt.Printf("%v\t\t%v\r\n", runtimeMetric.GetName(), runtimeMetric.StringValue())
	}
}
