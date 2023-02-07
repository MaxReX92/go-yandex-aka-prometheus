package storage

import (
	"fmt"
	"go-yandex-aka-prometheus/internal/logger"
	"go-yandex-aka-prometheus/internal/metrics"
	"strings"
	"sync"
)

type inMemoryStorage struct {
	gaugeMetrics   map[string]metrics.Metric
	counterMetrics map[string]metrics.Metric
	lock           sync.RWMutex
}

func NewInMemoryStorage() MetricsStorage {
	return &inMemoryStorage{
		gaugeMetrics:   map[string]metrics.Metric{},
		counterMetrics: map[string]metrics.Metric{},
		lock:           sync.RWMutex{},
	}
}

func (s *inMemoryStorage) AddGaugeMetricValue(name string, value float64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ensureMetricUpdate(s.counterMetrics, name, value, metrics.NewGaugeMetric)
}

func (s *inMemoryStorage) AddCounterMetricValue(name string, value int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ensureMetricUpdate(s.counterMetrics, name, float64(value), metrics.NewCounterMetric)
}

func (s *inMemoryStorage) GetMetrics() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	builder := strings.Builder{}
	for key, metric := range s.gaugeMetrics {
		builder.WriteString(fmt.Sprintf("%v: %v\r\n", key, metric.StringValue()))
	}

	for key, metric := range s.counterMetrics {
		builder.WriteString(fmt.Sprintf("%v: %v\r\n", key, metric.StringValue()))
	}

	return builder.String()
}

func ensureMetricUpdate(metricsMap map[string]metrics.Metric, name string, value float64, metricFactory func(string) metrics.Metric) {
	currentMetric, ok := metricsMap[name]
	if !ok {
		currentMetric := metricFactory(name)
		metricsMap[name] = currentMetric
	}

	currentMetric.SetValue(value)
	logger.InfoFormat("Updated metric: %v. value: %v", name, currentMetric.StringValue())
}
