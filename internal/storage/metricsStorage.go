package storage

type MetricsStorage interface {
	AddGaugeMetricValue(name string, value float64) float64
	AddCounterMetricValue(name string, value int64) int64
	GetMetricValues() map[string]map[string]string
	GetMetricValue(metricType string, metricName string) (float64, bool)

	Flush() error
	Restore(rawMetrics string)
	Close()
}
