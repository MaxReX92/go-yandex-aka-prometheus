package provider

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type aggregateMetricsProvider struct {
	providers []metrics.MetricsProvider
}

func NewAggregateMetricsProvider(providers ...metrics.MetricsProvider) *aggregateMetricsProvider {
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
	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < len(a.providers); i++ {
		num := i
		eg.Go(func() error {
			err := a.providers[num].Update(ctx)
			if err != nil {
				return logger.WrapError("update metrics", err)
			}

			return nil
		})
	}

	return eg.Wait()
}
