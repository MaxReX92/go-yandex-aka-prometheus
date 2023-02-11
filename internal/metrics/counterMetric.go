package metrics

import (
	"go-yandex-aka-prometheus/internal/parser"
)

type counterMetric struct {
	name  string
	value int64
}

func NewCounterMetric(name string) Metric {
	return &counterMetric{
		name:  name,
		value: 0,
	}
}

func (m *counterMetric) GetType() string {
	return "counter"
}

func (m *counterMetric) GetName() string {
	return m.name
}

func (m *counterMetric) GetStringValue() string {
	return parser.IntToString(m.value)
}

func (m *counterMetric) SetValue(value float64) {
	m.value = m.value + int64(value)
}

func (m *counterMetric) Flush() {
	m.value = 0
}
