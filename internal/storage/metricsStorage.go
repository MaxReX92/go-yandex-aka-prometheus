package storage

import "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"

type MetricsStorage interface {
	AddGaugeMetricValue(name string, value float64) (float64, error)
	AddCounterMetricValue(name string, value int64) (int64, error)
	AddMetricValue(metric metrics.Metric) (metrics.Metric, error)
	GetMetricValues() (map[string]map[string]string, error)
	GetMetricValue(metricType string, metricName string) (float64, error)

	Restore(metricValues map[string]map[string]string) error
}
