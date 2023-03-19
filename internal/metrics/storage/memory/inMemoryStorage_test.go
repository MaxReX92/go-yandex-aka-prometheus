package memory

import (
	"context"
	"errors"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

func TestInMemoryStorage_AddCounterMetricValue(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []test.KeyValue
		expected       map[string]map[string]string
	}{
		{
			name: "single_metric",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100}},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "100"}},
		}, {
			name: "single_negative_metric",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: -100}},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "-100"}},
		}, {
			name: "multi_metrics",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100},
				{Key: "metricName2", Value: 200},
			},
			expected: map[string]map[string]string{
				"counter": {
					"metricName1": "100",
					"metricName2": "200",
				}},
		},
		{
			name: "same_metrics",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100},
				{Key: "metricName1", Value: 200},
			},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "300"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			metricsList := make([]metrics.Metric, len(tt.counterMetrics))
			for i, m := range tt.counterMetrics {
				metricsList[i] = test.CreateCounterMetric(m.Key, m.Value)
			}
			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			actual, _ := storage.GetMetricValues(context.Background())
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_AddGaugeMetricValue(t *testing.T) {
	tests := []struct {
		name         string
		gaugeMetrics []test.KeyValue
		expected     map[string]map[string]string
	}{
		{
			name: "single_metric",
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100.001}},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "100.001"}},
		}, {
			name: "single_negative_metric",
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName1", Value: -100.001}},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "-100.001"}},
		}, {
			name: "multi_metrics",
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100.001},
				{Key: "metricName2", Value: 200.002},
			},
			expected: map[string]map[string]string{
				"gauge": {
					"metricName1": "100.001",
					"metricName2": "200.002",
				}},
		},
		{
			name: "same_metrics",
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100.001},
				{Key: "metricName1", Value: 200.002},
			},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "200.002"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			metricsList := make([]metrics.Metric, len(tt.gaugeMetrics))
			for i, m := range tt.gaugeMetrics {
				metricsList[i] = test.CreateGaugeMetric(m.Key, m.Value)
			}
			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			actual, _ := storage.GetMetricValues(context.Background())
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_GetMetricValues(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []test.KeyValue
		gaugeMetrics   []test.KeyValue
		expected       map[string]map[string]string
	}{
		{
			name:     "no_metric",
			expected: map[string]map[string]string{},
		}, {
			name: "all_metric",
			counterMetrics: []test.KeyValue{
				{Key: "metricName2", Value: 300},
				{Key: "metricName1", Value: 100},
				{Key: "metricName3", Value: -400},
				{Key: "metricName1", Value: 200}},
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName5", Value: 300.003},
				{Key: "metricName4", Value: 100.001},
				{Key: "metricName6", Value: -400.004},
				{Key: "metricName4", Value: 200.002}},
			expected: map[string]map[string]string{
				"counter": {
					"metricName1": "300",
					"metricName2": "300",
					"metricName3": "-400",
				},
				"gauge": {
					"metricName4": "200.002",
					"metricName5": "300.003",
					"metricName6": "-400.004",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			metricsList := make([]metrics.Metric, len(tt.counterMetrics)+len(tt.gaugeMetrics))
			for i, m := range tt.counterMetrics {
				metricsList[i] = test.CreateCounterMetric(m.Key, m.Value)
			}
			for i, m := range tt.gaugeMetrics {
				metricsList[len(tt.counterMetrics)+i] = test.CreateGaugeMetric(m.Key, m.Value)
			}
			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			actual, _ := storage.GetMetricValues(context.Background())
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_Restore(t *testing.T) {

	tests := []struct {
		name          string
		values        map[string]map[string]string
		expectedError error
	}{
		{
			name:          "unknown_metric_type",
			expectedError: errors.New("unknown metric types from backup: unknownType"),
			values: map[string]map[string]string{
				"unknownType": {
					"metricName1": "300",
				},
			},
		},
		{
			name: "success",
			values: map[string]map[string]string{
				"counter": {
					"metricName1": "300",
					"metricName2": "300",
					"metricName3": "-400",
				},
				"gauge": {
					"metricName4": "200.002",
					"metricName5": "300.003",
					"metricName6": "-400.004",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			actualError := storage.Restore(context.Background(), tt.values)
			assert.Equal(t, tt.expectedError, actualError)

			if tt.expectedError == nil {
				actual, _ := storage.GetMetricValues(context.Background())
				assert.Equal(t, tt.values, actual)
			}
		})
	}
}

func TestInMemoryStorage_GetMetricValue(t *testing.T) {
	tests := []struct {
		name             string
		counterMetrics   []test.KeyValue
		gaugeMetrics     []test.KeyValue
		expectedOk       bool
		expectedCounters []test.KeyValue
		expectedGauges   []test.KeyValue
	}{
		{
			name:             "empty_metrics",
			counterMetrics:   []test.KeyValue{},
			gaugeMetrics:     []test.KeyValue{},
			expectedOk:       false,
			expectedCounters: []test.KeyValue{{Key: "not_existed_metric", Value: 0}},
			expectedGauges:   []test.KeyValue{{Key: "not_existed_metric", Value: 0}},
		},
		{
			name: "metric_not_found",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100},
				{Key: "metricName2", Value: 300},
				{Key: "metricName3", Value: -400}},
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName4", Value: 100.001},
				{Key: "metricName5", Value: 300.003},
				{Key: "metricName6", Value: -400.004}},
			expectedOk:       false,
			expectedCounters: []test.KeyValue{{Key: "not_existed_metric", Value: 0}},
			expectedGauges:   []test.KeyValue{{Key: "not_existed_metric", Value: 0}},
		},
		{
			name: "success_values",
			counterMetrics: []test.KeyValue{
				{Key: "metricName1", Value: 100},
				{Key: "metricName2", Value: 300},
				{Key: "metricName3", Value: -400}},
			gaugeMetrics: []test.KeyValue{
				{Key: "metricName4", Value: 100.001},
				{Key: "metricName5", Value: 300.003},
				{Key: "metricName6", Value: -400.004}},
			expectedOk: true,
			expectedCounters: []test.KeyValue{
				{Key: "metricName1", Value: 100},
				{Key: "metricName2", Value: 300},
				{Key: "metricName3", Value: -400}},
			expectedGauges: []test.KeyValue{
				{Key: "metricName4", Value: 100.001},
				{Key: "metricName5", Value: 300.003},
				{Key: "metricName6", Value: -400.004}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			metricsList := make([]metrics.Metric, len(tt.counterMetrics)+len(tt.gaugeMetrics))
			for i, m := range tt.counterMetrics {
				metricsList[i] = test.CreateCounterMetric(m.Key, m.Value)
			}
			for i, m := range tt.gaugeMetrics {
				metricsList[len(tt.counterMetrics)+i] = test.CreateGaugeMetric(m.Key, m.Value)
			}

			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			for _, expectedCounter := range tt.expectedCounters {
				actualValue, err := storage.GetMetric(context.Background(), "counter", expectedCounter.Key)
				if tt.expectedOk {
					assert.NoError(t, err)
					assert.Equal(t, expectedCounter.Value, actualValue.GetValue())
				} else {
					assert.Error(t, err)
				}
			}

			for _, expectedGauge := range tt.expectedGauges {
				actualValue, err := storage.GetMetric(context.Background(), "gauge", expectedGauge.Key)
				if tt.expectedOk {
					assert.NoError(t, err)
					assert.Equal(t, expectedGauge.Value, actualValue.GetValue())
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}
