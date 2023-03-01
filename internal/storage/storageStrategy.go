package storage

import (
	"sync"
	"time"
)

type storageStrategyConfig interface {
	StoreInterval() time.Duration
}

type storageStrategy struct {
	inMemoryStorage MetricsStorage
	fileStorage     MetricsStorage
	storeInterval   time.Duration
	lock            sync.RWMutex
}

func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) MetricsStorage {
	return &storageStrategy{
		storeInterval:   config.StoreInterval(),
		inMemoryStorage: inMemoryStorage,
		fileStorage:     fileStorage,
	}
}

func (s *storageStrategy) AddGaugeMetricValue(name string, value float64) float64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := s.inMemoryStorage.AddGaugeMetricValue(name, value)
	if s.isSyncMode() {
		s.fileStorage.AddGaugeMetricValue(name, value)
	}

	return result
}

func (s *storageStrategy) AddCounterMetricValue(name string, value int64) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := s.inMemoryStorage.AddCounterMetricValue(name, value)
	if s.isSyncMode() {
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

func (s *storageStrategy) isSyncMode() bool {
	return s.storeInterval == 0
}
