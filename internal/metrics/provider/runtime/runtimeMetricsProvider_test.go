package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type config struct {
	metricNames []string
}

func (c *config) MetricsList() []string {
	return c.metricNames
}

func TestRuntimeMetricsProvider_Update(t *testing.T) {
	type expected struct {
		expectError   bool
		expectMetrics []string
	}

	tests := []struct {
		name        string
		metricNames []string
		expected    expected
	}{
		{
			name:        "empty_metrics_list",
			metricNames: []string{},
			expected: expected{
				expectError:   false,
				expectMetrics: []string{},
			},
		}, {
			name:        "unknown_metric_name",
			metricNames: []string{"UnknownMetricName"},
			expected: expected{
				expectError:   true,
				expectMetrics: []string{},
			},
		}, {
			name:        "correct_metrics_list",
			metricNames: []string{"Alloc", "LastGC"},
			expected: expected{
				expectError:   false,
				expectMetrics: []string{"Alloc", "LastGC"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			provider := NewRuntimeMetricsProvider(&config{metricNames: tt.metricNames})
			err := provider.Update(ctx)

			if tt.expected.expectError {
				assert.Error(t, err)
			} else {
				actualMetrics := test.ChanToArray(provider.GetMetrics())
				assert.Equal(t, len(tt.expected.expectMetrics), len(actualMetrics))
				for _, actualMetric := range actualMetrics {
					assert.Contains(t, tt.expected.expectMetrics, actualMetric.GetName())
				}
			}
		})
	}
}

func TestRuntimeMetricsProvider_GetMetrics(t *testing.T) {
	expectedMetrics := []string{"Alloc", "TotalAlloc"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	provider := NewRuntimeMetricsProvider(&config{metricNames: expectedMetrics})
	assert.NoErrorf(t, provider.Update(ctx), "fail to update metrics")

	actualMetrics := test.ChanToArray(provider.GetMetrics())
	assert.Len(t, actualMetrics, len(expectedMetrics))
	for _, actualMetric := range actualMetrics {
		assert.Contains(t, expectedMetrics, actualMetric.GetName())
		assert.NotEqual(t, actualMetric.GetStringValue(), "0")
	}
}
