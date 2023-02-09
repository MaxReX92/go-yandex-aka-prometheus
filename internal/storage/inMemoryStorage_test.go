package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type val interface {
	int64 | float64
}

type keyValue[T val] struct {
	key   string
	value T
}

func TestInMemoryStorage_AddCounterMetricValue(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []keyValue[int64]
		expected       string
	}{
		{
			name: "single_metric",
			counterMetrics: []keyValue[int64]{
				{key: "metricName1", value: 100}},
			expected: "metricName1: 100\r\n",
		}, {
			name: "single_negative_metric",
			counterMetrics: []keyValue[int64]{
				{key: "metricName1", value: -100}},
			expected: "metricName1: -100\r\n",
		}, {
			name: "multi_metrics",
			counterMetrics: []keyValue[int64]{
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 200},
			},
			expected: "metricName1: 100\r\nmetricName2: 200\r\n",
		},
		{
			name: "same_metrics",
			counterMetrics: []keyValue[int64]{
				{key: "metricName1", value: 100},
				{key: "metricName1", value: 200},
			},
			expected: "metricName1: 300\r\n",
		},
		{
			name: "order_metrics",
			counterMetrics: []keyValue[int64]{
				{key: "metricName3", value: 300},
				{key: "metricName1", value: 100},
				{key: "metricName2", value: 200},
			},
			expected: "metricName1: 100\r\n" +
				"metricName2: 200\r\n" +
				"metricName3: 300\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			for _, m := range tt.counterMetrics {
				storage.AddCounterMetricValue(m.key, m.value)
			}

			actual := storage.GetMetrics()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_AddGaugeMetricValue(t *testing.T) {
	tests := []struct {
		name         string
		gaugeMetrics []keyValue[float64]
		expected     string
	}{
		{
			name: "single_metric",
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName1", value: 100.001}},
			expected: "metricName1: 100.001\r\n",
		}, {
			name: "single_negative_metric",
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName1", value: -100.001}},
			expected: "metricName1: -100.001\r\n",
		}, {
			name: "multi_metrics",
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName1", value: 100.001},
				{key: "metricName2", value: 200.002},
			},
			expected: "metricName1: 100.001\r\nmetricName2: 200.002\r\n",
		},
		{
			name: "same_metrics",
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName1", value: 100.001},
				{key: "metricName1", value: 200.002},
			},
			expected: "metricName1: 200.002\r\n",
		},
		{
			name: "order_metrics",
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName3", value: 300.003},
				{key: "metricName1", value: 100.001},
				{key: "metricName2", value: 200.002},
			},
			expected: "metricName1: 100.001\r\nmetricName2: 200.002\r\nmetricName3: 300.003\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			for _, m := range tt.gaugeMetrics {
				storage.AddGaugeMetricValue(m.key, m.value)
			}

			actual := storage.GetMetrics()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInMemoryStorage_GetMetrics(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics []keyValue[int64]
		gaugeMetrics   []keyValue[float64]
		expected       string
	}{
		{
			name:     "no_metric",
			expected: "",
		}, {
			name: "all_metric",
			counterMetrics: []keyValue[int64]{
				{key: "metricName2", value: 300},
				{key: "metricName1", value: 100},
				{key: "metricName3", value: -400},
				{key: "metricName1", value: 200}},
			gaugeMetrics: []keyValue[float64]{
				{key: "metricName5", value: 300.003},
				{key: "metricName4", value: 100.001},
				{key: "metricName6", value: -400.004},
				{key: "metricName4", value: 200.002}},
			expected: "metricName1: 300\r\n" +
				"metricName2: 300\r\n" +
				"metricName3: -400\r\n" +
				"metricName4: 200.002\r\n" +
				"metricName5: 300.003\r\n" +
				"metricName6: -400.004\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()

			for _, m := range tt.counterMetrics {
				storage.AddCounterMetricValue(m.key, m.value)
			}

			for _, m := range tt.gaugeMetrics {
				storage.AddGaugeMetricValue(m.key, m.value)
			}

			actual := storage.GetMetrics()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
