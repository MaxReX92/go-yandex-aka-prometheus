package types

import (
	"fmt"
	"hash"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type counterMetric struct {
	name         string
	value        int64
	getValueChan chan int64
	setValueChan chan int64
}

func NewCounterMetric(name string) metrics.Metric {
	counter := &counterMetric{
		name:         name,
		getValueChan: make(chan int64),
		setValueChan: make(chan int64),
	}

	go func(metric *counterMetric) {
		for {
			select {
			case metric.getValueChan <- metric.value:
			case metric.value = <-metric.setValueChan:
			}
		}

	}(counter)

	return counter
}

func (m *counterMetric) GetType() string {
	return "counter"
}

func (m *counterMetric) GetName() string {
	return m.name
}

func (m *counterMetric) GetValue() float64 {
	return float64(<-m.getValueChan)
}

func (m *counterMetric) GetStringValue() string {
	return parser.IntToString(<-m.getValueChan)
}

func (m *counterMetric) SetValue(value float64) float64 {
	return m.setValue(<-m.getValueChan + int64(value))
}

func (m *counterMetric) Flush() {
	m.setValue(0)
}

func (m *counterMetric) GetHash(hash hash.Hash) ([]byte, error) {
	value := <-m.getValueChan
	_, err := hash.Write([]byte(fmt.Sprintf("%s:counter:%d", m.name, value)))
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func (m *counterMetric) setValue(value int64) float64 {
	m.setValueChan <- value
	return float64(value)
}
