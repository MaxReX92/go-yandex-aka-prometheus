package storage

import "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"

type MetricsStorage interface {
	AddMetricValue(metric metrics.Metric) (metrics.Metric, error)
	GetMetricValues() (map[string]map[string]string, error)
	GetMetric(metricType string, metricName string) (metrics.Metric, error)
	Restore(metricValues map[string]map[string]string) error
}
