package metrics

import (
	"context"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"math/rand"
)

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

func (c *CustomMetricsProvider) GetMetrics(context.Context) []Metric {
	return []Metric{
		&c.poolMetric,
		&c.randomMetric,
	}
}

func (c *CustomMetricsProvider) Update(context.Context) error {
	logger.Info("Start collect custom metrics")

	c.poolMetric.SetValue(1)
	logger.InfoFormat("Updated metric: %v. value: %v", c.poolMetric.GetName(), c.poolMetric.StringValue())

	c.randomMetric.SetValue(rand.Float64())
	logger.InfoFormat("Updated metric: %v. value: %v", c.randomMetric.GetName(), c.randomMetric.StringValue())

	return nil
}
