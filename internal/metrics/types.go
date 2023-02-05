package metrics

type Metric interface {
	GetName() string
	GetType() string
	StringValue() string
	SetValue(value float64)
}

type MetricsProvider interface {
	GetMetrics() []Metric
	Update() error
}
