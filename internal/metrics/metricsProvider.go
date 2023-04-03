package metrics

import (
	"context"
)

type MetricsProvider interface {
	GetMetrics() <-chan (Metric)
	Update(ctx context.Context) error
}
