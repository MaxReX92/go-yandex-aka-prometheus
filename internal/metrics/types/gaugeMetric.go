package types

import (
	"fmt"
	"hash"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type gaugeMetric struct {
	name  string
	value float64
	lock  sync.RWMutex
}

func NewGaugeMetric(name string) metrics.Metric {
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

func (m *gaugeMetric) GetValue() float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.value
}

func (m *gaugeMetric) GetStringValue() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return parser.FloatToString(m.value)
}

func (m *gaugeMetric) SetValue(value float64) float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = value
	return m.value
}

func (m *gaugeMetric) Flush() {
}

func (m *gaugeMetric) GetHash(hash hash.Hash) ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, err := hash.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.name, m.value)))
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
