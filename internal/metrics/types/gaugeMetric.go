package types

import (
	"fmt"
	"hash"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type gaugeMetric struct {
	name         string
	value        float64
	getValueChan chan float64
	setValueChan chan float64
}

func NewGaugeMetric(name string) metrics.Metric {
	gauge := &gaugeMetric{
		name:         name,
		getValueChan: make(chan float64),
		setValueChan: make(chan float64),
	}

	go func(metric *gaugeMetric) {
		for {
			select {
			case metric.getValueChan <- metric.value:
			case metric.value = <-metric.setValueChan:
			}
		}

	}(gauge)

	return gauge
}

func (m *gaugeMetric) GetType() string {
	return "gauge"
}

func (m *gaugeMetric) GetName() string {
	return m.name
}

func (m *gaugeMetric) GetValue() float64 {
	return <-m.getValueChan
}

func (m *gaugeMetric) GetStringValue() string {
	return parser.FloatToString(m.GetValue())
}

func (m *gaugeMetric) SetValue(value float64) float64 {
	m.setValueChan <- value
	return value
}

func (m *gaugeMetric) Flush() {
}

func (m *gaugeMetric) GetHash(hash hash.Hash) ([]byte, error) {
	value := <-m.getValueChan
	_, err := hash.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.name, value)))
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
