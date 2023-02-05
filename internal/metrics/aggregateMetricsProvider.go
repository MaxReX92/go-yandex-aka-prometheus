package metrics

type AggregateMetricsProvider struct {
	providers []MetricsProvider
}

func NewAggregateMetricsProvider(providers []MetricsProvider) AggregateMetricsProvider {
	return AggregateMetricsProvider{
		providers: providers,
	}
}

func (a *AggregateMetricsProvider) GetMetrics() []Metric {
	resultMetrics := []Metric{}
	for _, provider := range a.providers {
		resultMetrics = append(resultMetrics, provider.GetMetrics()...)
	}

	return resultMetrics
}

func (a *AggregateMetricsProvider) Update() error {
	for _, provider := range a.providers {
		err := provider.Update()
		if err != nil {
			return err
		}
	}

	return nil
}
