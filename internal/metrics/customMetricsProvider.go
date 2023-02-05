package metrics

import "math/rand"

type CustomMetricsProvider struct {
	poolMetric   CounterMetric
	randomMetric GaugeMetric
}

func NewCustomMetricsProvider() CustomMetricsProvider {
	return CustomMetricsProvider{
		poolMetric: CounterMetric{
			name:  "PollCount",
			value: 0,
		},
		randomMetric: GaugeMetric{
			name:  "RandomValue",
			value: 0,
		},
	}
}

func (c CustomMetricsProvider) GetMetrics() []Metric {
	return []Metric{
		&c.poolMetric,
		&c.randomMetric,
	}
}

func (c CustomMetricsProvider) Update() error {
	c.poolMetric.SetValue(1)
	c.randomMetric.SetValue(rand.Float64())
	return nil
}
