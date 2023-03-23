package metrics

import "github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"

type Metric interface {
	hash.HashHolder

	GetName() string
	GetType() string
	GetValue() float64
	GetStringValue() string
	SetValue(value float64) float64
	Flush()
}
