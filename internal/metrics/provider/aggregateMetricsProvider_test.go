package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type aggregateMetricsProviderMock struct {
	mock.Mock
}

func TestAggregateMetricsProvider_GetMetrics(t *testing.T) {
	counter := types.NewCounterMetric("counterMetric")
	gauge := types.NewCounterMetric("gaugeMetric")

	tests := []struct {
		name                  string
		firstProviderMetrics  []metrics.Metric
		secondProviderMetrics []metrics.Metric
		expectedMetrics       []metrics.Metric
	}{
		{
			name:                  "empty_metrics",
			firstProviderMetrics:  []metrics.Metric{},
			secondProviderMetrics: []metrics.Metric{},
			expectedMetrics:       []metrics.Metric{},
		},
		{
			name:                  "success",
			firstProviderMetrics:  []metrics.Metric{counter},
			secondProviderMetrics: []metrics.Metric{gauge},
			expectedMetrics:       []metrics.Metric{counter, gauge},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstProvider := new(aggregateMetricsProviderMock)
			secondProvider := new(aggregateMetricsProviderMock)

			firstProvider.On("GetMetrics").Return(tt.firstProviderMetrics)
			secondProvider.On("GetMetrics").Return(tt.secondProviderMetrics)

			provider := NewAggregateMetricsProvider(firstProvider, secondProvider)
			actualMetrics := provider.GetMetrics()

			assert.Equal(t, tt.expectedMetrics, actualMetrics)

			firstProvider.AssertCalled(t, "GetMetrics")
			secondProvider.AssertCalled(t, "GetMetrics")
		})
	}
}

func TestAggregateMetricsProvider_Update(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                string
		firstProviderError  error
		secondProviderError error
		expectedError       error
	}{
		{
			name:               "first_provider_error",
			firstProviderError: test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:                "second_provider_error",
			secondProviderError: test.ErrTest,
			expectedError:       test.ErrTest,
		},
		{
			name: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstProvider := new(aggregateMetricsProviderMock)
			secondProvider := new(aggregateMetricsProviderMock)

			firstProvider.On("Update", ctx).Return(tt.firstProviderError)
			secondProvider.On("Update", ctx).Return(tt.secondProviderError)

			provider := NewAggregateMetricsProvider(firstProvider, secondProvider)
			actualError := provider.Update(ctx)

			assert.ErrorIs(t, actualError, tt.expectedError)

			firstProvider.AssertCalled(t, "Update", ctx)
			if tt.firstProviderError == nil {
				secondProvider.AssertCalled(t, "Update", ctx)
			} else {
				secondProvider.AssertNotCalled(t, "Update", mock.Anything)
			}
		})
	}
}

func (a *aggregateMetricsProviderMock) GetMetrics() []metrics.Metric {
	args := a.Called()
	return args.Get(0).([]metrics.Metric)
}

func (a *aggregateMetricsProviderMock) Update(ctx context.Context) error {
	args := a.Called(ctx)
	return args.Error(0)
}

func (a *aggregateMetricsProviderMock) GetMetricsChan() <-chan metrics.Metric {
	args := a.Called()
	return args.Get(0).(<-chan metrics.Metric)
}
