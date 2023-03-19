package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

type runtimeMetricsProviderConfig interface {
	MetricsList() []string
}

type runtimeMetricsProvider struct {
	metrics []metrics.Metric
}

func NewRuntimeMetricsProvider(config runtimeMetricsProviderConfig) metrics.MetricsProvider {
	metricNames := config.MetricsList()
	metricsList := make([]metrics.Metric, len(metricNames))
	for i, metricName := range metricNames {
		metricsList[i] = types.NewGaugeMetric(metricName)
	}

	return &runtimeMetricsProvider{metrics: metricsList}
}

func (p *runtimeMetricsProvider) Update(context.Context) error {
	logger.Info("Start collect runtime metrics")
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	for _, metric := range p.metrics {
		metricName := metric.GetName()
		metricValue, err := getFieldValue(&stats, metricName)
		if err != nil {
			return logger.WrapError(fmt.Sprintf("get %s runtime metric value", metricName), err)
		}

		metric.SetValue(metricValue)
		logger.InfoFormat("Updated metric: %v. value: %v", metricName, metric.GetStringValue())
	}

	return nil
}

func (p *runtimeMetricsProvider) GetMetrics() []metrics.Metric {
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
