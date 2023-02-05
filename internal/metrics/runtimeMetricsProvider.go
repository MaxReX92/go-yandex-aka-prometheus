package metrics

import (
	"errors"
	"reflect"
	"runtime"
)

type RuntimeMetricsProviderConfig struct {
	MetricsList []string
}

type RuntimeMetricsProvider struct {
	metrics []Metric
}

func NewRuntimeMetricsProvider(config RuntimeMetricsProviderConfig) RuntimeMetricsProvider {
	metrics := []Metric{}
	for _, metricName := range config.MetricsList {
		metrics = append(metrics, &GaugeMetric{
			name:  metricName,
			value: 0,
		})
	}

	return RuntimeMetricsProvider{metrics: metrics}
}

func (p RuntimeMetricsProvider) Update() error {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	for _, metric := range p.metrics {
		metricValue, err := getFieldValue(&stats, metric.GetName())
		if err != nil {
			return err
		}

		metric.SetValue(metricValue)
	}

	return nil
}

func (p RuntimeMetricsProvider) GetMetrics() []Metric {
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

	return 0, errors.New("Unknown value type" + value.Type().Name())
}
