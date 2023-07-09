package http

import (
	"fmt"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

// MetricsConverterConfig contains required metrics converter settings.
type MetricsConverterConfig interface {
	SignMetrics() bool
}

// Converter provides model converter functionality.
type Converter struct {
	signer      *hash.Signer
	signMetrics bool
}

// NewMetricsConverter create new instance of Converter.
func NewMetricsConverter(conf MetricsConverterConfig, signer *hash.Signer) *Converter {
	return &Converter{
		signMetrics: conf.SignMetrics(),
		signer:      signer,
	}
}

// ToModelMetric convert internal dsl metric to model metric.
func (c *Converter) ToModelMetric(metric metrics.Metric) (*model.Metrics, error) {
	modelMetric := &model.Metrics{
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
		return nil, logger.WrapError(fmt.Sprintf("convert metric with type %s", modelMetric.MType), metrics.ErrUnknownMetricType)
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

// FromModelMetric convert model metric to internal dsl metric.
func (c *Converter) FromModelMetric(modelMetric *model.Metrics) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.MType {
	case "counter":
		if modelMetric.Delta == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewCounterMetric(modelMetric.ID)
		value = float64(*modelMetric.Delta)
	case "gauge":
		if modelMetric.Value == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewGaugeMetric(modelMetric.ID)
		value = *modelMetric.Value
	default:

		return nil, logger.WrapError(fmt.Sprintf("convert metric with type %s", modelMetric.MType), metrics.ErrUnknownMetricType)
	}

	metric.SetValue(value)

	if c.signMetrics && modelMetric.Hash != "" {
		ok, err := c.signer.CheckSign(metric, modelMetric.Hash)
		if err != nil {
			return nil, logger.WrapError("check signature", err)
		}

		if !ok {
			return nil, logger.WrapError("check signature", metrics.ErrInvalidSignature)
		}
	}

	return metric, nil
}
