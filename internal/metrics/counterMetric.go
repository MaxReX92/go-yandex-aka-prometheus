package metrics

import (
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type counterMetric struct {
	name  string
	value int64
	lock  sync.RWMutex
}

func NewCounterMetric(name string) Metric {
	return &counterMetric{
		name: name,
	}
}

func (m *counterMetric) GetType() string {
	return "counter"
}

func (m *counterMetric) GetName() string {
	return m.name
}

func (m *counterMetric) GetValue() float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return float64(m.value)
}

func (m *counterMetric) GetStringValue() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return parser.IntToString(m.value)
}

func (m *counterMetric) SetValue(value float64) float64 {
	return m.setValue(m.value + int64(value))
}

func (m *counterMetric) Flush() {
	m.setValue(0)
}

func (m *counterMetric) setValue(value int64) float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = value
	return float64(m.value)
}
