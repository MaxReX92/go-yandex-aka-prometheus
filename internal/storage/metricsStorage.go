package storage

import (
	"fmt"
	"strings"
	"sync"
)

type MetricsStorage struct {
	gaugeMetrics   map[string]string
	counterMetrics map[string]string
	lock           sync.RWMutex
}

func NewMetricsStorage() MetricsStorage {
	return MetricsStorage{
		gaugeMetrics:   map[string]string{},
		counterMetrics: map[string]string{},
		lock:           sync.RWMutex{},
	}
}

func (s *MetricsStorage) AddGaugeMetricValue(name string, stringValue string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.gaugeMetrics[name] = stringValue
}

func (s *MetricsStorage) AddCounterMetricValue(name string, stringValue string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.counterMetrics[name] = stringValue
}

func (s *MetricsStorage) GetMetrics() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	builder := strings.Builder{}

	for _, metricsList := range []map[string]string{s.gaugeMetrics, s.counterMetrics} {
		for key, value := range metricsList {
			builder.WriteString(fmt.Sprintf("%v: %v\r\n", key, value))
		}
	}

	return builder.String()
}
