package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomMetricsProvider_GetMetrics(t *testing.T) {
	provider := NewCustomMetricsProvider()

	metrics := provider.GetMetrics()
	assert.Len(t, metrics, 2)

	poolMetric := metrics[0]
	assert.Equal(t, "PollCount", poolMetric.GetName())
	assert.Equal(t, "counter", poolMetric.GetType())
	assert.Equal(t, float64(0), poolMetric.GetValue())

	randomMetric := metrics[1]
	assert.Equal(t, "RandomValue", randomMetric.GetName())
	assert.Equal(t, "gauge", randomMetric.GetType())
	assert.Equal(t, float64(0), randomMetric.GetValue())
}

func TestCustomMetricsProvider_Update(t *testing.T) {
	ctx := context.Background()
	provider := NewCustomMetricsProvider()

	metrics := provider.GetMetrics()
	assert.Len(t, metrics, 2)

	poolMetric := metrics[0]
	randomMetric := metrics[1]

	assert.Equal(t, float64(0), poolMetric.GetValue())
	assert.Equal(t, float64(0), randomMetric.GetValue())

	err := provider.Update(ctx)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), poolMetric.GetValue())
	assert.NotEqual(t, float64(0), randomMetric.GetValue())

	randomValue := randomMetric.GetValue()
	err = provider.Update(ctx)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), poolMetric.GetValue())
	assert.NotEqual(t, randomValue, randomMetric.GetValue())
}
