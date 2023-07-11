package grpc

import (
	"fmt"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"github.com/MaxReX92/go-yandex-aka-prometheus/proto/generated"
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
func (c *Converter) ToModelMetric(metric metrics.Metric) (*generated.Metric, error) {
	modelMetric := &generated.Metric{
		Name: metric.GetName(),
	}

	metricType := metric.GetType()
	metricValue := metric.GetValue()
	switch metricType {
	case "counter":
		counterValue := int64(metricValue)
		modelMetric.Delta = &counterValue
		modelMetric.Type = generated.MetricType_COUNTER
	case "gauge":
		modelMetric.Value = &metricValue
		modelMetric.Type = generated.MetricType_GAUGE
	default:
		return nil, logger.WrapError(fmt.Sprintf("convert metric with type %s", metricType), metrics.ErrUnknownMetricType)
	}

	if c.signMetrics {
		signature, err := c.signer.GetSign(metric)
		if err != nil {
			return nil, logger.WrapError("get signature", err)
		}

		modelMetric.Hash = signature
	}

	return modelMetric, nil
}

// FromModelMetric convert model metric to internal dsl metric.
func (c *Converter) FromModelMetric(modelMetric *generated.Metric) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.Type {
	case generated.MetricType_COUNTER:
		if modelMetric.Delta == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewCounterMetric(modelMetric.Name)
		value = float64(*modelMetric.Delta)
	case generated.MetricType_GAUGE:
		if modelMetric.Value == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewGaugeMetric(modelMetric.Name)
		value = *modelMetric.Value
	default:

		return nil, logger.WrapError(fmt.Sprintf("convert metric with type %s", modelMetric.Type), metrics.ErrUnknownMetricType)
	}

	metric.SetValue(value)

	if c.signMetrics && modelMetric.Hash != nil {
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
