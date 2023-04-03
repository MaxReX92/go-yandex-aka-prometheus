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

			firstProvider.On("GetMetrics").Return(test.ArrayToChan(tt.firstProviderMetrics))
			secondProvider.On("GetMetrics").Return(test.ArrayToChan(tt.secondProviderMetrics))

			provider := NewAggregateMetricsProvider(firstProvider, secondProvider)
			actualMetrics := test.ChanToArray(provider.GetMetrics())

			assert.ElementsMatch(t, tt.expectedMetrics, actualMetrics)

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

			firstProvider.On("Update", mock.Anything).Return(tt.firstProviderError)
			secondProvider.On("Update", mock.Anything).Return(tt.secondProviderError)

			provider := NewAggregateMetricsProvider(firstProvider, secondProvider)
			actualError := provider.Update(ctx)

			assert.ErrorIs(t, actualError, tt.expectedError)

			firstProvider.AssertCalled(t, "Update", mock.Anything)
			secondProvider.AssertCalled(t, "Update", mock.Anything)
		})
	}
}

func (a *aggregateMetricsProviderMock) Update(ctx context.Context) error {
	args := a.Called(ctx)
	return args.Error(0)
}

func (a *aggregateMetricsProviderMock) GetMetrics() <-chan metrics.Metric {
	args := a.Called()
	return args.Get(0).(<-chan metrics.Metric)
}
