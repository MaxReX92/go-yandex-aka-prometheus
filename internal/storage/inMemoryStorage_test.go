package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type keyValue struct {
	key   string
	value float64
}

func TestInMemoryStorage_AddCounterMetricValue(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []keyValue
		expected       map[string]map[string]string
	}{
		{
			name: "single_metric",
			counterMetrics: []keyValue{
				{key: "metricName1", value: 100}},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "100"}},
		}, {
			name: "single_negative_metric",
			counterMetrics: []keyValue{
				{key: "metricName1", value: -100}},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "-100"}},
		}, {
			name: "multi_metrics",
			counterMetrics: []keyValue{
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 200},
			},
			expected: map[string]map[string]string{
				"counter": {
					"metricName1": "100",
					"metricName2": "200",
				}},
		},
		{
			name: "same_metrics",
			counterMetrics: []keyValue{
				{key: "metricName1", value: 100},
				{key: "metricName1", value: 200},
			},
			expected: map[string]map[string]string{
				"counter": {"metricName1": "300"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			for _, m := range tt.counterMetrics {
				_, err := storage.AddMetricValue(createCounterMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			actual, _ := storage.GetMetricValues()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_AddGaugeMetricValue(t *testing.T) {
	tests := []struct {
		name         string
		gaugeMetrics []keyValue
		expected     map[string]map[string]string
	}{
		{
			name: "single_metric",
			gaugeMetrics: []keyValue{
				{key: "metricName1", value: 100.001}},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "100.001"}},
		}, {
			name: "single_negative_metric",
			gaugeMetrics: []keyValue{
				{key: "metricName1", value: -100.001}},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "-100.001"}},
		}, {
			name: "multi_metrics",
			gaugeMetrics: []keyValue{
				{key: "metricName1", value: 100.001},
				{key: "metricName2", value: 200.002},
			},
			expected: map[string]map[string]string{
				"gauge": {
					"metricName1": "100.001",
					"metricName2": "200.002",
				}},
		},
		{
			name: "same_metrics",
			gaugeMetrics: []keyValue{
				{key: "metricName1", value: 100.001},
				{key: "metricName1", value: 200.002},
			},
			expected: map[string]map[string]string{
				"gauge": {"metricName1": "200.002"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			for _, m := range tt.gaugeMetrics {
				_, err := storage.AddMetricValue(createGaugeMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			actual, _ := storage.GetMetricValues()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_GetMetricValues(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []keyValue
		gaugeMetrics   []keyValue
		expected       map[string]map[string]string
	}{
		{
			name:     "no_metric",
			expected: map[string]map[string]string{},
		}, {
			name: "all_metric",
			counterMetrics: []keyValue{
				{key: "metricName2", value: 300},
				{key: "metricName1", value: 100},
				{key: "metricName3", value: -400},
				{key: "metricName1", value: 200}},
			gaugeMetrics: []keyValue{
				{key: "metricName5", value: 300.003},
				{key: "metricName4", value: 100.001},
				{key: "metricName6", value: -400.004},
				{key: "metricName4", value: 200.002}},
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

			for _, m := range tt.counterMetrics {
				_, err := storage.AddMetricValue(createCounterMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			for _, m := range tt.gaugeMetrics {
				_, err := storage.AddMetricValue(createGaugeMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			actual, _ := storage.GetMetricValues()
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
			expectedError: errors.New("unknown metric type from backup: unknownType"),
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

			actualError := storage.Restore(tt.values)
			assert.Equal(t, tt.expectedError, actualError)

			if tt.expectedError == nil {
				actual, _ := storage.GetMetricValues()
				assert.Equal(t, tt.values, actual)
			}
		})
	}
}

func TestInMemoryStorage_GetMetricValue(t *testing.T) {
	tests := []struct {
		name             string
		counterMetrics   []keyValue
		gaugeMetrics     []keyValue
		expectedOk       bool
		expectedCounters []keyValue
		expectedGauges   []keyValue
	}{
		{
			name:             "empty_metrics",
			counterMetrics:   []keyValue{},
			gaugeMetrics:     []keyValue{},
			expectedOk:       false,
			expectedCounters: []keyValue{{key: "not_existed_metric", value: 0}},
			expectedGauges:   []keyValue{{key: "not_existed_metric", value: 0}},
		},
		{
			name: "metric_not_found",
			counterMetrics: []keyValue{
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 300},
				{key: "metricName3", value: -400}},
			gaugeMetrics: []keyValue{
				{key: "metricName4", value: 100.001},
				{key: "metricName5", value: 300.003},
				{key: "metricName6", value: -400.004}},
			expectedOk:       false,
			expectedCounters: []keyValue{{key: "not_existed_metric", value: 0}},
			expectedGauges:   []keyValue{{key: "not_existed_metric", value: 0}},
		},
		{
			name: "success_values",
			counterMetrics: []keyValue{
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 300},
				{key: "metricName3", value: -400}},
			gaugeMetrics: []keyValue{
				{key: "metricName4", value: 100.001},
				{key: "metricName5", value: 300.003},
				{key: "metricName6", value: -400.004}},
			expectedOk: true,
			expectedCounters: []keyValue{
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 300},
				{key: "metricName3", value: -400}},
			expectedGauges: []keyValue{
				{key: "metricName4", value: 100.001},
				{key: "metricName5", value: 300.003},
				{key: "metricName6", value: -400.004}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			for _, m := range tt.counterMetrics {
				_, err := storage.AddMetricValue(createCounterMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			for _, m := range tt.gaugeMetrics {
				_, err := storage.AddMetricValue(createGaugeMetric(m.key, m.value))
				assert.NoError(t, err)
			}

			for _, expectedCounter := range tt.expectedCounters {
				actualValue, err := storage.GetMetric("counter", expectedCounter.key)
				if tt.expectedOk {
					assert.NoError(t, err)
					assert.Equal(t, expectedCounter.value, actualValue.GetValue())
				} else {
					assert.Error(t, err)
				}
			}

			for _, expectedGauge := range tt.expectedGauges {
				actualValue, err := storage.GetMetric("gauge", expectedGauge.key)
				if tt.expectedOk {
					assert.NoError(t, err)
					assert.Equal(t, expectedGauge.value, actualValue.GetValue())
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}
