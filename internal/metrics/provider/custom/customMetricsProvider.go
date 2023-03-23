package custom

import (
	"context"
	"math/rand"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

type customMetricsProvider struct {
	poolMetric   metrics.Metric
	randomMetric metrics.Metric
}

func NewCustomMetricsProvider() metrics.MetricsProvider {
	return &customMetricsProvider{
		poolMetric:   types.NewCounterMetric("PollCount"),
		randomMetric: types.NewGaugeMetric("RandomValue"),
	}
}

func (c *customMetricsProvider) GetMetrics() []metrics.Metric {
	return []metrics.Metric{
		c.poolMetric,
		c.randomMetric,
	}
}

func (c *customMetricsProvider) Update(context.Context) error {
	logger.Info("Start collect custom metrics")

	c.poolMetric.SetValue(1)
	logger.InfoFormat("Updated metric: %v. value: %v", c.poolMetric.GetName(), c.poolMetric.GetStringValue())

	c.randomMetric.SetValue(rand.Float64())
	logger.InfoFormat("Updated metric: %v. value: %v", c.randomMetric.GetName(), c.randomMetric.GetStringValue())

	return nil
}
