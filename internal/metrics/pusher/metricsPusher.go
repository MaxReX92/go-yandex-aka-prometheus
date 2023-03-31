package pusher

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type MetricsPusher interface {
	Push(ctx context.Context, metrics []metrics.Metric) error
	PushChan(ctx context.Context, metrics <-chan metrics.Metric) error
}
