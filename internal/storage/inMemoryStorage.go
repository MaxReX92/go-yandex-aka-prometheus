package storage

import (
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type inMemoryStorage struct {
	metricsByType map[string]map[string]metrics.Metric
	lock          sync.RWMutex
}

func NewInMemoryStorage() MetricsStorage {
	return &inMemoryStorage{
		metricsByType: map[string]map[string]metrics.Metric{},
		lock:          sync.RWMutex{},
	}
}

func (s *inMemoryStorage) AddGaugeMetricValue(name string, value float64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ensureMetricUpdate("gauge", name, value, metrics.NewGaugeMetric)
}

func (s *inMemoryStorage) AddCounterMetricValue(name string, value int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ensureMetricUpdate("counter", name, float64(value), metrics.NewCounterMetric)
}

func (s *inMemoryStorage) GetMetricValues() map[string]map[string]string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metricValues := map[string]map[string]string{}
	for metricsType, metricsList := range s.metricsByType {
		values := map[string]string{}
		metricValues[metricsType] = values

		for metricName, metric := range metricsList {
			values[metricName] = metric.GetStringValue()
		}
	}

	return metricValues
}

func (s *inMemoryStorage) GetMetricValue(metricType string, metricName string) (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metricsByName, ok := s.metricsByType[metricType]
	if !ok {
		logger.ErrorFormat("Metrics with type %v not found", metricType)
		return "", false
	}

	metric, ok := metricsByName[metricName]
	if !ok {
		logger.ErrorFormat("Metrics with name %v and type %v not found", metricName, metricType)
		return "", false
	}

	return metric.GetStringValue(), true
}

func (s *inMemoryStorage) ensureMetricUpdate(metricType string, name string, value float64, metricFactory func(string) metrics.Metric) {
	metricsList, ok := s.metricsByType[metricType]
	if !ok {
		metricsList = map[string]metrics.Metric{}
		s.metricsByType[metricType] = metricsList
	}

	currentMetric, ok := metricsList[name]
	if !ok {
		currentMetric = metricFactory(name)
		metricsList[name] = currentMetric
	}

	currentMetric.SetValue(value)
}
