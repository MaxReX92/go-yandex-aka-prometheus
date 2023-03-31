package provider

import (
	"context"
	"sync"

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

func (a *aggregateMetricsProvider) GetMetrics() <-chan metrics.Metric {
	result := make(chan metrics.Metric)

	go func() {
		wg := sync.WaitGroup{}
		for _, provider := range a.providers {
			wg.Add(1)
			go func(p metrics.MetricsProvider) {
				for metric := range p.GetMetrics() {
					result <- metric
				}
				wg.Done()
			}(provider)
		}

		wg.Wait()
		close(result)
	}()

	return result
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
