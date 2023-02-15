package metrics

import "context"

type MetricsProvider interface {
	GetMetrics() []Metric
	Update(ctx context.Context) error
}
