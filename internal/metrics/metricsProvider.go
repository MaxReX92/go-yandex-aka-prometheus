package metrics

import (
	"context"
)

// MetricsProvider work with metrics.
type MetricsProvider interface {
	// GetMetrics return all provider metrics.
	GetMetrics() <-chan (Metric)

	// Update method update all provider metrics.
	Update(ctx context.Context) error
}
