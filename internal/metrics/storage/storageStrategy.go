package storage

import (
	"context"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type storageStrategyConfig interface {
	SyncMode() bool
}

// StorageStrategy combine in memory and long-term metrics storages.
type StorageStrategy struct {
	backupStorage   MetricsStorage
	inMemoryStorage MetricsStorage
	syncMode        bool
	lock            sync.RWMutex
}

// NewStorageStrategy creates new instance of StorageStrategy.
func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) *StorageStrategy {
	return &StorageStrategy{
		backupStorage:   fileStorage,
		inMemoryStorage: inMemoryStorage,
		syncMode:        config.SyncMode(),
	}
}

func (s *StorageStrategy) AddMetricValues(ctx context.Context, metric []metrics.Metric) ([]metrics.Metric, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result, err := s.inMemoryStorage.AddMetricValues(ctx, metric)
	if err != nil {
		return result, logger.WrapError("add metric values to memory storage", err)
	}

	if s.syncMode {
		_, err = s.backupStorage.AddMetricValues(ctx, result)
		if err != nil {
			return nil, logger.WrapError("add metric values to backup storage", err)
		}
	}

	return result, nil
}

func (s *StorageStrategy) GetMetricValues(ctx context.Context) (map[string]map[string]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetricValues(ctx)
}

func (s *StorageStrategy) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inMemoryStorage.GetMetric(ctx, metricType, metricName)
}

func (s *StorageStrategy) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.inMemoryStorage.Restore(ctx, metricValues)
}

func (s *StorageStrategy) CreateBackup(ctx context.Context) error {
	currentState, err := s.inMemoryStorage.GetMetricValues(ctx)
	if err != nil {
		return logger.WrapError("get metrics from memory storage", err)
	}

	return s.backupStorage.Restore(ctx, currentState)
}

func (s *StorageStrategy) RestoreFromBackup(ctx context.Context) error {
	restoredState, err := s.backupStorage.GetMetricValues(ctx)
	if err != nil {
		return logger.WrapError("get metrics from backup storage", err)
	}

	return s.inMemoryStorage.Restore(ctx, restoredState)
}

func (s *StorageStrategy) Close() error {
	return s.CreateBackup(context.Background()) // force backup
}
