package gopsutil

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

type GopsutilMetricsProvider struct {
	totalMetric           metrics.Metric
	freeMetric            metrics.Metric
	cpuUtilizationMetrics map[int]metrics.Metric
}

func NewGopsutilMetricsProvider() *GopsutilMetricsProvider {
	numCPU := runtime.NumCPU()
	cpuUtilizationMetrics := make(map[int]metrics.Metric, numCPU)
	for i := 0; i < numCPU; i++ {
		cpuUtilizationMetrics[i] = types.NewGaugeMetric(fmt.Sprintf("CPUutilization%v", i+1))
	}

	return &GopsutilMetricsProvider{
		totalMetric:           types.NewCounterMetric("TotalMemory"),
		freeMetric:            types.NewGaugeMetric("FreeMemory"),
		cpuUtilizationMetrics: cpuUtilizationMetrics,
	}
}

func (g *GopsutilMetricsProvider) GetMetrics() <-chan metrics.Metric {
	result := make(chan metrics.Metric)
	go func() {
		defer close(result)
		result <- g.totalMetric
		result <- g.freeMetric
		for _, metric := range g.cpuUtilizationMetrics {
			result <- metric
		}
	}()

	return result
}

func (g *GopsutilMetricsProvider) Update(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return g.updateMemoryMetrics(ctx) })
	eg.Go(func() error { return g.updateCPUMetrics(ctx) })

	return eg.Wait()
}

func (g *GopsutilMetricsProvider) updateMemoryMetrics(ctx context.Context) error {
	memoryStats, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return logger.WrapError("get memory stats", err)
	}

	g.totalMetric.SetValue(float64(memoryStats.Total))
	logger.InfoFormat("Updated metric: %v. value: %v", g.totalMetric.GetName(), g.totalMetric.GetStringValue())

	g.freeMetric.SetValue(float64(memoryStats.Free))
	logger.InfoFormat("Updated metric: %v. value: %v", g.freeMetric.GetName(), g.freeMetric.GetStringValue())

	return nil
}

func (g *GopsutilMetricsProvider) updateCPUMetrics(ctx context.Context) error {
	cpuStats, err := cpu.PercentWithContext(ctx, time.Second, true)
	if err != nil {
		return logger.WrapError("get cpu stats", err)
	}

	for i, val := range cpuStats {
		metric := g.cpuUtilizationMetrics[i]
		metric.SetValue(val)
		logger.InfoFormat("Updated metric: %v. value: %v", metric.GetName(), metric.GetStringValue())
	}

	return nil
}
