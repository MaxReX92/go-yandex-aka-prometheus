package storage

import (
	"context"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type MetricsStorage interface {
	AddMetricValue(ctx context.Context, metric metrics.Metric) (metrics.Metric, error)
	GetMetricValues(ctx context.Context) (map[string]map[string]string, error)
	GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error)
	Restore(ctx context.Context, metricValues map[string]map[string]string) error
}
