package metrics

import (
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type gaugeMetric struct {
	name  string
	value float64
	lock  sync.RWMutex
}

func NewGaugeMetric(name string) Metric {
	return &gaugeMetric{
		name: name,
	}
}

func (m *gaugeMetric) GetType() string {
	return "gauge"
}

func (m *gaugeMetric) GetName() string {
	return m.name
}

func (m *gaugeMetric) GetStringValue() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return parser.FloatToString(m.value)
}

func (m *gaugeMetric) SetValue(value float64) float64 {
	return m.setValue(value)
}

func (m *gaugeMetric) Flush() {
}

func (m *gaugeMetric) setValue(value float64) float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = value
	return m.value
}
