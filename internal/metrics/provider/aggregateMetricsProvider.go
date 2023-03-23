package provider

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type aggregateMetricsProvider struct {
	providers []metrics.MetricsProvider
}

func NewAggregateMetricsProvider(providers ...metrics.MetricsProvider) metrics.MetricsProvider {
	return &aggregateMetricsProvider{
		providers: providers,
	}
}

func (a *aggregateMetricsProvider) GetMetrics() []metrics.Metric {
	resultMetrics := []metrics.Metric{}
	for _, provider := range a.providers {
		resultMetrics = append(resultMetrics, provider.GetMetrics()...)
	}

	return resultMetrics
}

func (a *aggregateMetricsProvider) Update(ctx context.Context) error {
	for _, provider := range a.providers {
		err := provider.Update(ctx)
		if err != nil {
			return logger.WrapError("update metrics", err)
		}
	}

	return nil
}
