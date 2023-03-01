package storage

import "time"

type storageStrategyConfig interface {
	StoreInterval() time.Duration
}

type storageStrategy struct {
	inMemoryStorage MetricsStorage
	fileStorage     MetricsStorage
	storeInterval   time.Duration
}

func NewStorageStrategy(config storageStrategyConfig, inMemoryStorage MetricsStorage, fileStorage MetricsStorage) MetricsStorage {
	return &storageStrategy{
		storeInterval:   config.StoreInterval(),
		inMemoryStorage: inMemoryStorage,
		fileStorage:     fileStorage,
	}
}

func (s *storageStrategy) AddGaugeMetricValue(name string, value float64) float64 {
	//TODO implement me
	panic("implement me")
}

func (s *storageStrategy) AddCounterMetricValue(name string, value int64) int64 {
	//TODO implement me
	panic("implement me")
}

func (s *storageStrategy) GetMetricValues() map[string]map[string]string {
	//TODO implement me
	panic("implement me")
}

func (s *storageStrategy) GetMetricValue(metricType string, metricName string) (float64, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *storageStrategy) Restore(rawMetrics string) {
	//TODO implement me
	panic("implement me")
}
