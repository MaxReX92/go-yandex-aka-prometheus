package html

import (
	"github.com/stretchr/testify/assert"
	"go-yandex-aka-prometheus/internal/metrics"
	"testing"
)

func TestSimplePageBuilder_BuildMetricsPage(t *testing.T) {
	tests := []struct {
		name           string
		counterMetrics map[string]metrics.Metric
		gaugeMetrics   map[string]metrics.Metric
		expected       string
	}{
		{
			name:           "no_metric",
			counterMetrics: map[string]metrics.Metric{},
			gaugeMetrics:   map[string]metrics.Metric{},
			expected:       "<html></html>",
		}, {
			name: "all_metric",
			counterMetrics: map[string]metrics.Metric{
				"metricName2": createCounterMetric("metricName2", 300),
				"metricName3": createCounterMetric("metricName3", -400),
				"metricName1": createCounterMetric("metricName1", 200)},
			gaugeMetrics: map[string]metrics.Metric{
				"metricName5": createGaugeMetric("metricName5", 300.003),
				"metricName4": createGaugeMetric("metricName4", 100.001),
				"metricName6": createGaugeMetric("metricName6", -400.004)},
			expected: "<html>" +
				"metricName1: 200<br>" +
				"metricName2: 300<br>" +
				"metricName3: -400<br>" +
				"metricName4: 100.001<br>" +
				"metricName5: 300.003<br>" +
				"metricName6: -400.004<br>" +
				"</html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSimplePageBuilder()
			metricsByType := map[string]map[string]metrics.Metric{
				"counter": tt.counterMetrics,
				"gauge":   tt.gaugeMetrics,
			}

			actual := builder.BuildMetricsPage(metricsByType)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func createCounterMetric(name string, value int64) metrics.Metric {
	metric := metrics.NewCounterMetric(name)
	metric.SetValue(float64(value))
	return metric
}

func createGaugeMetric(name string, value float64) metrics.Metric {
	metric := metrics.NewGaugeMetric(name)
	metric.SetValue(value)
	return metric
}
