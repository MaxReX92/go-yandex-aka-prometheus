package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type aggregateMetricsProviderMock struct {
	mock.Mock
}

func TestAggregateMetricsProvider_GetMetrics(t *testing.T) {

	counter := NewCounterMetric("counterMetric")
	gauge := NewCounterMetric("gaugeMetric")

	tests := []struct {
		name                  string
		firstProviderMetrics  []Metric
		secondProviderMetrics []Metric
		expectedMetrics       []Metric
	}{
		{
			name:                  "empty_metrics",
			firstProviderMetrics:  []Metric{},
			secondProviderMetrics: []Metric{},
			expectedMetrics:       []Metric{},
		},
		{
			name:                  "success",
			firstProviderMetrics:  []Metric{counter},
			secondProviderMetrics: []Metric{gauge},
			expectedMetrics:       []Metric{counter, gauge},
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
	testError := errors.New("testError")

	tests := []struct {
		name                string
		firstProviderError  error
		secondProviderError error
		expectedError       error
	}{
		{
			name:               "first_provider_error",
			firstProviderError: testError,
			expectedError:      testError,
		},
		{
			name:                "second_provider_error",
			secondProviderError: testError,
			expectedError:       testError,
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

			assert.Equal(t, tt.expectedError, actualError)

			firstProvider.AssertCalled(t, "Update", ctx)
			if tt.firstProviderError == nil {
				secondProvider.AssertCalled(t, "Update", ctx)
			} else {
				secondProvider.AssertNotCalled(t, "Update", mock.Anything)
			}
		})
	}
}

func (a *aggregateMetricsProviderMock) GetMetrics() []Metric {
	args := a.Called()
	return args.Get(0).([]Metric)
}

func (a *aggregateMetricsProviderMock) Update(ctx context.Context) error {
	args := a.Called(ctx)
	return args.Error(0)
}
