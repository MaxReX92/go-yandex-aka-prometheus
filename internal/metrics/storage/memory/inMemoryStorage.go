package memory

import (
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type inMemoryStorage struct {
	metricsByType map[string]map[string]metrics.Metric
	lock          sync.RWMutex
}

func NewInMemoryStorage() storage.MetricsStorage {
	return &inMemoryStorage{
		metricsByType: map[string]map[string]metrics.Metric{},
		lock:          sync.RWMutex{},
	}
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

func (s *inMemoryStorage) GetMetric(metricType string, metricName string) (metrics.Metric, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metricsByName, ok := s.metricsByType[metricType]
	if !ok {
		return nil, fmt.Errorf("metrics with types %v not found", metricType)
	}

	metric, ok := metricsByName[metricName]
	if !ok {
		return nil, fmt.Errorf("metrics with name %v and types %v not found", metricName, metricType)
	}

	return metric, nil
}

func (s *inMemoryStorage) Restore(metricValues map[string]map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.metricsByType = map[string]map[string]metrics.Metric{}

	for metricType, metricsByType := range metricValues {
		metricFactory := types.NewGaugeMetric
		if metricType == "counter" {
			metricFactory = types.NewCounterMetric
		} else if metricType != "gauge" {
			return fmt.Errorf("unknown metric types from backup: %v", metricType)
		}

		for metricName, metricValue := range metricsByType {
			value, err := parser.ToFloat64(metricValue)
			if err != nil {
				return err
			}

			metricsList, ok := s.metricsByType[metricType]
			if !ok {
				metricsList = map[string]metrics.Metric{}
				s.metricsByType[metricType] = metricsList
			}

			currentMetric, ok := metricsList[metricName]
			if !ok {
				currentMetric = metricFactory(metricName)
				metricsList[metricName] = currentMetric
			}

			currentMetric.SetValue(value)
		}
	}

	return nil
}
