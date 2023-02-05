package metrics

import "fmt"

type GaugeMetric struct {
	name  string
	value float64
}

func (m *GaugeMetric) GetType() string {
	return "gauge"
}

func (m *GaugeMetric) GetName() string {
	return m.name
}

func (m *GaugeMetric) StringValue() string {
	return fmt.Sprintf("%g", m.value)
}

func (m *GaugeMetric) SetValue(value float64) {
	m.value = value
}
