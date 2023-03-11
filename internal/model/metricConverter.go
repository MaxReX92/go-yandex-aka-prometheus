package model

import (
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

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
		return nil, fmt.Errorf("unknown metric type: %v", modelMetric.MType)
	}

	return modelMetric, nil
}

func FromModelMetric(modelMetric *Metrics) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.MType {
	case "counter":
		metric = metrics.NewCounterMetric(modelMetric.ID)
		value = float64(*modelMetric.Delta)
	case "gauge":
		metric = metrics.NewGaugeMetric(modelMetric.ID)
		value = *modelMetric.Value
	default:
		return nil, fmt.Errorf("unknown metric type: %v", modelMetric.MType)
	}

	metric.SetValue(value)
	return metric, nil
}
