package storage

import (
	"fmt"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
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

func (s *inMemoryStorage) AddGaugeMetricValue(name string, value float64) (float64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ensureMetricUpdate("gauge", name, value, metrics.NewGaugeMetric), nil
}

func (s *inMemoryStorage) AddCounterMetricValue(name string, value int64) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return int64(s.ensureMetricUpdate("counter", name, float64(value), metrics.NewCounterMetric)), nil
}

func (s *inMemoryStorage) AddMetricValue(metric metrics.Metric) (metrics.Metric, error) {
	metricType := metric.GetType()
	metricsList, ok := s.metricsByType[metricType]
	if !ok {
		metricsList = map[string]metrics.Metric{}
		s.metricsByType[metricType] = metricsList
	}

	metricName := metric.GetName()
	currentMetric, ok := metricsList[metricName]
	if ok {
		currentMetric.SetValue(metric.GetValue())
	} else {
		currentMetric = metric
		metricsList[metricName] = currentMetric
	}

	return currentMetric, nil
}

func (s *inMemoryStorage) GetMetricValues() (map[string]map[string]string, error) {
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

	return metricValues, nil
}

func (s *inMemoryStorage) GetMetricValue(metricType string, metricName string) (float64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metricsByName, ok := s.metricsByType[metricType]
	if !ok {
		return 0, fmt.Errorf("metrics with type %v not found", metricType)
	}

	metric, ok := metricsByName[metricName]
	if !ok {
		return 0, fmt.Errorf("metrics with name %v and type %v not found", metricName, metricType)
	}

	return metric.GetValue(), nil
}

func (s *inMemoryStorage) Restore(metricValues map[string]map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.metricsByType = map[string]map[string]metrics.Metric{}

	for metricType, metricsByType := range metricValues {
		metricFactory := metrics.NewGaugeMetric
		if metricType == "counter" {
			metricFactory = metrics.NewCounterMetric
		} else if metricType != "gauge" {
			return fmt.Errorf("unknown metric type from backup: %v", metricType)
		}

		for metricName, metricValue := range metricsByType {
			value, err := parser.ToFloat64(metricValue)
			if err != nil {
				return err
			}
			s.ensureMetricUpdate(metricType, metricName, value, metricFactory)
		}
	}

	return nil
}

func (s *inMemoryStorage) ensureMetricUpdate(metricType string, name string, value float64, metricFactory func(string) metrics.Metric) float64 {
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

	return currentMetric.SetValue(value)
}
