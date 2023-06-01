package pusher

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

// MetricsPusher send metrics to remote storage.
type MetricsPusher interface {
	// Push method send served metrics to remote storage.
	Push(ctx context.Context, metrics <-chan metrics.Metric) error
}
