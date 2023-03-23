package test

import (
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

type KeyValue struct {
	Key   string
	Value float64
}

func CreateCounterMetric(name string, value float64) metrics.Metric {
	return CreateMetric(types.NewCounterMetric, name, value)
}

func CreateGaugeMetric(name string, value float64) metrics.Metric {
	return CreateMetric(types.NewGaugeMetric, name, value)
}

func CreateMetric(metricFactory func(string) metrics.Metric, name string, value float64) metrics.Metric {
	metric := metricFactory(name)
	metric.SetValue(value)
	return metric
}
