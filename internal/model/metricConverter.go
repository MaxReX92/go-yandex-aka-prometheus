package model

import (
	"errors"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

var ErrUnknownMetricType = errors.New("unknown metric type")

func ToModelMetric(metric metrics.Metric) (*Metrics, error) {
	modelMetric := &Metrics{
		ID:    metric.GetName(),
		MType: metric.GetType(),
	}

	metricValue := metric.GetValue()
	switch modelMetric.MType {
	case "counter":
		counterValue := int64(metricValue)
		modelMetric.Delta = &counterValue
	case "gauge":
		modelMetric.Value = &metricValue
	default:
		logger.ErrorFormat("unknown metric type: %v", modelMetric.MType)
		return nil, ErrUnknownMetricType
	}

	return modelMetric, nil
}

func FromModelMetric(modelMetric *Metrics) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.MType {
	case "counter":
		if modelMetric.Delta == nil {
			return nil, errors.New("metric value is missed")
		}

		metric = metrics.NewCounterMetric(modelMetric.ID)
		value = float64(*modelMetric.Delta)
	case "gauge":
		if modelMetric.Value == nil {
			return nil, errors.New("metric value is missed")
		}

		metric = metrics.NewGaugeMetric(modelMetric.ID)
		value = *modelMetric.Value
	default:
		logger.ErrorFormat("unknown metric type: %v", modelMetric.MType)
		return nil, ErrUnknownMetricType
	}

	metric.SetValue(value)
	return metric, nil
}
