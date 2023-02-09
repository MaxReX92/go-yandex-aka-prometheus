package metrics

import "context"

type MetricsProvider interface {
	GetMetrics(ctx context.Context) []Metric
	Update(ctx context.Context) error
}
