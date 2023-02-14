package metrics

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

type aggregateMetricsProvider struct {
	providers []MetricsProvider
}

func NewAggregateMetricsProvider(providers []MetricsProvider) MetricsProvider {
	return &aggregateMetricsProvider{
		providers: providers,
	}
}

func (a *aggregateMetricsProvider) GetMetrics() []Metric {
	resultMetrics := []Metric{}
	for _, provider := range a.providers {
		resultMetrics = append(resultMetrics, provider.GetMetrics()...)
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
