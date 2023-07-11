package server

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type RequestHandler interface {
	GetMetricValue(ctx context.Context, metricType string, metricName string) (metrics.Metric, error)
	UpdateMetricValues(ctx context.Context, metricValues []metrics.Metric) ([]metrics.Metric, error)

	GetReportPage(ctx context.Context) (string, error)
	Ping(ctx context.Context) error
}
