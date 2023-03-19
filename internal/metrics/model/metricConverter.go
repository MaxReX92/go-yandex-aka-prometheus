package model

import (
	"errors"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

type ErrUnknownMetricType struct {
	UnknownType string
}

func (e *ErrUnknownMetricType) Error() string {
	return "unknown metric type: " + e.UnknownType
}

type MetricsConverterConfig interface {
	SignMetrics() bool
}

type MetricsConverter struct {
	signer      *hash.Signer
	signMetrics bool
}

func NewMetricsConverter(conf MetricsConverterConfig, signer *hash.Signer) *MetricsConverter {
	return &MetricsConverter{
		signMetrics: conf.SignMetrics(),
		signer:      signer,
	}
}

func (c *MetricsConverter) ToModelMetric(metric metrics.Metric) (*Metrics, error) {
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
		return nil, &ErrUnknownMetricType{UnknownType: modelMetric.MType}
	}

	if c.signMetrics {
		signature, err := c.signer.GetSignString(metric)
		if err != nil {
			return nil, logger.WrapError("get signature string", err)
		}

		modelMetric.Hash = signature
	}

	return modelMetric, nil
}

func (c *MetricsConverter) FromModelMetric(modelMetric *Metrics) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.MType {
	case "counter":
		if modelMetric.Delta == nil {
			return nil, errors.New("metric value is missed")
		}

		metric = types.NewCounterMetric(modelMetric.ID)
		value = float64(*modelMetric.Delta)
	case "gauge":
		if modelMetric.Value == nil {
			return nil, errors.New("metric value is missed")
		}

		metric = types.NewGaugeMetric(modelMetric.ID)
		value = *modelMetric.Value
	default:
		logger.ErrorFormat("unknown metric type: %v", modelMetric.MType)
		return nil, &ErrUnknownMetricType{UnknownType: modelMetric.MType}
	}

	metric.SetValue(value)

	if c.signMetrics && modelMetric.Hash != "" {
		ok, err := c.signer.CheckSign(metric, modelMetric.Hash)
		if err != nil {
			return nil, logger.WrapError("check signature", err)
		}

		if !ok {
			return nil, errors.New("invalid signature")
		}
	}

	return metric, nil
}
