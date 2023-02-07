package storage

type MetricsStorage interface {
	AddGaugeMetricValue(name string, value float64)
	AddCounterMetricValue(name string, value int64)
	GetMetrics() string
}
