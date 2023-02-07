package metrics

import (
	"context"
	"go-yandex-aka-prometheus/internal/logger"
)

type aggregateMetricsProvider struct {
	providers []MetricsProvider
}

func NewAggregateMetricsProvider(providers []MetricsProvider) MetricsProvider {
	return &aggregateMetricsProvider{
		providers: providers,
	}
}

func (a *aggregateMetricsProvider) GetMetrics(ctx context.Context) []Metric {
	resultMetrics := []Metric{}
	for _, provider := range a.providers {
		resultMetrics = append(resultMetrics, provider.GetMetrics(ctx)...)
	}

	return resultMetrics
}

func (a *aggregateMetricsProvider) Update(ctx context.Context) error {
	for _, provider := range a.providers {
		err := provider.Update(ctx)
		if err != nil {
			logger.ErrorFormat("Fail to update metrics: %v", err.Error())
			return err
		}
	}

	return nil
}
