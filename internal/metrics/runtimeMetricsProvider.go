package metrics

import (
	"context"
	"fmt"
	"go-yandex-aka-prometheus/internal/logger"
	"reflect"
	"runtime"
)

type RuntimeMetricsProviderConfig interface {
	MetricsList() []string
}

type runtimeMetricsProvider struct {
	metrics []Metric
}

func NewRuntimeMetricsProvider(config RuntimeMetricsProviderConfig) MetricsProvider {
	metrics := []Metric{}
	for _, metricName := range config.MetricsList() {
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
		logger.InfoFormat("Updated metric: %v. value: %v", metricName, metric.GetStringValue())
	}

	return nil
}

func (p *runtimeMetricsProvider) GetMetrics() []Metric {
	return p.metrics
}

func getFieldValue(stats *runtime.MemStats, fieldName string) (float64, error) {
	r := reflect.ValueOf(stats)
	f := reflect.Indirect(r).FieldByName(fieldName)

	value, ok := convertValue(f)
	if !ok {
		return value, fmt.Errorf("field name %v was not found", fieldName)
	}

	return value, nil
}

func convertValue(value reflect.Value) (float64, bool) {
	if value.CanInt() {
		return float64(value.Int()), true
	}
	if value.CanUint() {
		return float64(value.Uint()), true
	}
	if value.CanFloat() {
		return value.Float(), true
	}

	return 0, false
}
