package client

import (
	"context"
	"go-yandex-aka-prometheus/internal/metrics"
)

type MetricsPusher interface {
	Push(ctx context.Context, metrics []metrics.Metric) error
}
