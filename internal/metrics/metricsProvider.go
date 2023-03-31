package metrics

import (
	"context"
)

type MetricsProvider interface {
	GetMetricsChan() <-chan (Metric)
	GetMetrics() []Metric
	Update(ctx context.Context) error
}
