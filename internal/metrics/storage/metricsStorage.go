package storage

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

// MetricsStorage is a long-term storage of received metrics.
type MetricsStorage interface {
	// AddMetricValues serve metrics.
	AddMetricValues(ctx context.Context, metric []metrics.Metric) ([]metrics.Metric, error)

	// GetMetricValues returns all metric values.
	GetMetricValues(ctx context.Context) (map[string]map[string]string, error)

	// GetMetric returns single metric by metric type and name.
	GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error)

	// Restore recovers storage state.
	Restore(ctx context.Context, metricValues map[string]map[string]string) error
}
