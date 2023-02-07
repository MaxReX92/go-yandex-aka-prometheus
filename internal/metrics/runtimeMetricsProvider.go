package metrics

import (
	"context"
	"errors"
	"go-yandex-aka-prometheus/internal/logger"
	"reflect"
	"runtime"
)

type RuntimeMetricsProviderConfig struct {
	MetricsList []string
}

type runtimeMetricsProvider struct {
	metrics []Metric
}

func NewRuntimeMetricsProvider(config RuntimeMetricsProviderConfig) MetricsProvider {
	metrics := []Metric{}
	for _, metricName := range config.MetricsList {
		metrics = append(metrics, NewGaugeMetric(metricName))
	}

	return &runtimeMetricsProvider{metrics: metrics}
}

func (p *runtimeMetricsProvider) Update(context.Context) error {
	logger.Info("Start collect runtime metrics")
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	for _, metric := range p.metrics {
		metricName := metric.GetName()
		metricValue, err := getFieldValue(&stats, metricName)
		if err != nil {
			logger.ErrorFormat("Fail to get %v runtime metric value: %v", metricName, err.Error())
			return err
		}

		metric.SetValue(metricValue)
		logger.InfoFormat("Updated metric: %v. value: %v", metricName, metric.StringValue())
	}

	return nil
}

func (p *runtimeMetricsProvider) GetMetrics(context.Context) []Metric {
	return p.metrics
}

func getFieldValue(stats *runtime.MemStats, fieldName string) (float64, error) {
	r := reflect.ValueOf(stats)
	f := reflect.Indirect(r).FieldByName(fieldName)

	return convertValue(f)
}

func convertValue(value reflect.Value) (float64, error) {
	if value.CanInt() {
		return float64(value.Int()), nil
	}
	if value.CanUint() {
		return float64(value.Uint()), nil
	}
	if value.CanFloat() {
		return value.Float(), nil
	}

	return 0, errors.New("Unknown value type: " + value.Type().Name())
}
