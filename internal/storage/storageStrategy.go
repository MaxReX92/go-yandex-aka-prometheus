package storage

import (
	"sync"
)

type storageStrategyConfig interface {
	SyncMode() bool
}

type storageStrategy struct {
	fileStorage     MetricsStorage
	inMemoryStorage MetricsStorage
	syncMode        bool
	lock            sync.RWMutex
}

func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) MetricsStorage {
	return &storageStrategy{
		fileStorage:     fileStorage,
		inMemoryStorage: inMemoryStorage,
		syncMode:        config.SyncMode(),
	}
}

func (s *storageStrategy) AddGaugeMetricValue(name string, value float64) float64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := s.inMemoryStorage.AddGaugeMetricValue(name, value)
	if s.syncMode {
		s.fileStorage.AddGaugeMetricValue(name, value)
	}

	return result
}

func (s *storageStrategy) AddCounterMetricValue(name string, value int64) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := s.inMemoryStorage.AddCounterMetricValue(name, value)
	if s.syncMode {
		s.fileStorage.AddCounterMetricValue(name, value)
	}

	return result
}

func (s *storageStrategy) GetMetricValues() map[string]map[string]string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValues()
}

func (s *storageStrategy) GetMetricValue(metricType string, metricName string) (float64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValue(metricType, metricName)
}

func (s *storageStrategy) Restore(rawMetrics string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.inMemoryStorage.Restore(rawMetrics)
}

func (s *storageStrategy) Close() {
	s.inMemoryStorage.Close()
	s.fileStorage.Close()
}
