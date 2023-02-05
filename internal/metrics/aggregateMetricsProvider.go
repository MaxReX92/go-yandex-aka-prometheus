package metrics

import "context"

type AggregateMetricsProvider struct {
	providers []MetricsProvider
}

func NewAggregateMetricsProvider(providers []MetricsProvider) AggregateMetricsProvider {
	return AggregateMetricsProvider{
		providers: providers,
	}
}

func (a *AggregateMetricsProvider) GetMetrics(context.Context) []Metric {
	resultMetrics := []Metric{}
	for _, provider := range a.providers {
		resultMetrics = append(resultMetrics, provider.GetMetrics(nil)...)
	}

	return resultMetrics
}

func (a *AggregateMetricsProvider) Update(ctx context.Context) error {
	for _, provider := range a.providers {
		err := provider.Update(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
