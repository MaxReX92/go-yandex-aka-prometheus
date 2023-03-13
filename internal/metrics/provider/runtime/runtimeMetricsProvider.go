package runtime

import (
	"context"
	"fmt"
	metrics2 "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"reflect"
	"runtime"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

type runtimeMetricsProviderConfig interface {
	MetricsList() []string
}

type runtimeMetricsProvider struct {
	metrics []metrics2.Metric
}

func NewRuntimeMetricsProvider(config runtimeMetricsProviderConfig) metrics2.MetricsProvider {
	metricsList := config.MetricsList()
	metrics := make([]metrics2.Metric, len(metricsList))
	for i, metricName := range metricsList {
		metrics[i] = types.NewGaugeMetric(metricName)
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
			logger.ErrorFormat("Fail to get %v runtime metric value: %v", metricName, err)
			return err
		}

		metric.SetValue(metricValue)
		logger.InfoFormat("Updated metric: %v. value: %v", metricName, metric.GetStringValue())
	}

	return nil
}

func (p *runtimeMetricsProvider) GetMetrics() []metrics2.Metric {
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
