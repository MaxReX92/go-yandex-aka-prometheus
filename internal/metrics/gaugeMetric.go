package metrics

import (
	"go-yandex-aka-prometheus/internal/parser"
)

type gaugeMetric struct {
	name  string
	value float64
}

func NewGaugeMetric(name string) Metric {
	return &gaugeMetric{
		name:  name,
		value: 0,
	}
}

func (m *gaugeMetric) GetType() string {
	return "gauge"
}

func (m *gaugeMetric) GetName() string {
	return m.name
}

func (m *gaugeMetric) GetStringValue() string {
	return parser.FloatToString(m.value)
}

func (m *gaugeMetric) SetValue(value float64) {
	m.value = value
}

func (m *gaugeMetric) Flush() {
}
