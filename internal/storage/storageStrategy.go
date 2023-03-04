package storage

import (
	"sync"
)

type storageStrategyConfig interface {
	SyncMode() bool
}

type StorageStrategy struct {
	fileStorage     MetricsStorage
	inMemoryStorage MetricsStorage
	syncMode        bool
	lock            sync.RWMutex
}

func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) *StorageStrategy {
	return &StorageStrategy{
		fileStorage:     fileStorage,
		inMemoryStorage: inMemoryStorage,
		syncMode:        config.SyncMode(),
	}
}

func (s *StorageStrategy) AddGaugeMetricValue(name string, value float64) (float64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result, err := s.inMemoryStorage.AddGaugeMetricValue(name, value)
	if err != nil {
		return result, err
	}

	if s.syncMode {
		_, err = s.fileStorage.AddGaugeMetricValue(name, result)
		if err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (s *StorageStrategy) AddCounterMetricValue(name string, value int64) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result, err := s.inMemoryStorage.AddCounterMetricValue(name, value)
	if err != nil {
		return result, err
	}

	if s.syncMode {
		_, err = s.fileStorage.AddCounterMetricValue(name, result)
		if err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (s *StorageStrategy) GetMetricValues() (map[string]map[string]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValues()
}

func (s *StorageStrategy) GetMetricValue(metricType string, metricName string) (float64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValue(metricType, metricName)
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

	return s.fileStorage.Restore(currentState)
}

func (s *StorageStrategy) RestoreFromBackup() error {
	restoredState, err := s.fileStorage.GetMetricValues()
	if err != nil {
		return err
	}

	return s.inMemoryStorage.Restore(restoredState)
}
