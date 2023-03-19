package memory

import (
	"context"
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
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

func (s *inMemoryStorage) AddMetricValues(ctx context.Context, metricList []metrics.Metric) ([]metrics.Metric, error) {
	result := make([]metrics.Metric, len(metricList))

	for i, metric := range metricList {
		metricType := metric.GetType()
		typedMetrics, ok := s.metricsByType[metricType]
		if !ok {
			typedMetrics = map[string]metrics.Metric{}
			s.metricsByType[metricType] = typedMetrics
		}

		metricName := metric.GetName()
		currentMetric, ok := typedMetrics[metricName]
		if ok {
			currentMetric.SetValue(metric.GetValue())
		} else {
			currentMetric = metric
			typedMetrics[metricName] = currentMetric
		}

		result[i] = currentMetric
	}

	return result, nil
}

func (s *inMemoryStorage) GetMetricValues(context.Context) (map[string]map[string]string, error) {
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

func (s *inMemoryStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
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

func (s *inMemoryStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
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
				return logger.WrapError("parse float metric value", err)
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
