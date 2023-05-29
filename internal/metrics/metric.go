package metrics

import "github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"

// Metric is a main abstraction of metric.
type Metric interface {
	hash.HashHolder

	// GetName returns metric name.
	GetName() string

	// GetType returns metric type.
	GetType() string

	// GetValue returns metric value.
	GetValue() float64

	// GetStringValue return metric value string representation.
	GetStringValue() string

	// SetValue updates metric value.
	SetValue(value float64) float64

	// Flush reset metric state, if needed,
	Flush()
}
