package storage

import (
	"fmt"
	"go-yandex-aka-prometheus/internal/metrics"
	"sort"
	"strings"
	"sync"
)

type inMemoryStorage struct {
	names          []string
	gaugeMetrics   map[string]metrics.Metric
	counterMetrics map[string]metrics.Metric
	lock           sync.RWMutex
}

func NewInMemoryStorage() MetricsStorage {
	return &inMemoryStorage{
		names:          []string{},
		gaugeMetrics:   map[string]metrics.Metric{},
		counterMetrics: map[string]metrics.Metric{},
		lock:           sync.RWMutex{},
	}
}

func (s *inMemoryStorage) AddGaugeMetricValue(name string, value float64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ensureMetricUpdate(s.gaugeMetrics, &s.names, name, value, metrics.NewGaugeMetric)
}

func (s *inMemoryStorage) AddCounterMetricValue(name string, value int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ensureMetricUpdate(s.counterMetrics, &s.names, name, float64(value), metrics.NewCounterMetric)
}

func (s *inMemoryStorage) GetMetrics() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	builder := strings.Builder{}
	for _, metricName := range s.names {
		metric, ok := s.counterMetrics[metricName]
		if ok {
			builder.WriteString(fmt.Sprintf("%v: %v\r\n", metricName, metric.StringValue()))
			continue
		}

		metric, ok = s.gaugeMetrics[metricName]
		if ok {
			builder.WriteString(fmt.Sprintf("%v: %v\r\n", metricName, metric.StringValue()))
		}
	}

	return builder.String()
}

func ensureMetricUpdate(metricsMap map[string]metrics.Metric, keys *[]string, name string, value float64, metricFactory func(string) metrics.Metric) {
	currentMetric, ok := metricsMap[name]
	if !ok {
		currentMetric = metricFactory(name)
		metricsMap[name] = currentMetric

		// need revision
		*keys = append(*keys, name)
		sort.Strings(*keys)
	}

	currentMetric.SetValue(value)
}
