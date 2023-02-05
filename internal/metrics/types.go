package metrics

import "context"

type Metric interface {
	GetName() string
	GetType() string
	StringValue() string
	SetValue(value float64)
}

type MetricsProvider interface {
	GetMetrics(ctx context.Context) []Metric
	Update(ctx context.Context) error
}
