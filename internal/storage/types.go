package storage

type MetricsStorage interface {
	AddGaugeMetricValue(name string, stringValue string)
	AddCounterMetricValue(name string, stringValue string)
	GetMetrics() string
}
