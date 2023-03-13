package storage

import (
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"sync"
)

type storageStrategyConfig interface {
	SyncMode() bool
}

type StorageStrategy struct {
	backupStorage   MetricsStorage
	inMemoryStorage MetricsStorage
	syncMode        bool
	lock            sync.RWMutex
}

func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) *StorageStrategy {
	return &StorageStrategy{
		backupStorage:   fileStorage,
		inMemoryStorage: inMemoryStorage,
		syncMode:        config.SyncMode(),
	}
}

func (s *StorageStrategy) AddMetricValue(metric metrics.Metric) (metrics.Metric, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result, err := s.inMemoryStorage.AddMetricValue(metric)
	if err != nil {
		return result, err
	}

	if s.syncMode {
		_, err = s.backupStorage.AddMetricValue(result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *StorageStrategy) GetMetricValues() (map[string]map[string]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValues()
}

func (s *StorageStrategy) GetMetric(metricType string, metricName string) (metrics.Metric, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetric(metricType, metricName)
}

func (s *StorageStrategy) Restore(metricValues map[string]map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.inMemoryStorage.Restore(metricValues)
}

func (s *StorageStrategy) CreateBackup() error {
	currentState, err := s.inMemoryStorage.GetMetricValues()
	if err != nil {
		return err
	}

	return s.backupStorage.Restore(currentState)
}

func (s *StorageStrategy) RestoreFromBackup() error {
	restoredState, err := s.backupStorage.GetMetricValues()
	if err != nil {
		return err
	}

	return s.inMemoryStorage.Restore(restoredState)
}
