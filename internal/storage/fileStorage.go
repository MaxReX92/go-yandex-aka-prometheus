package storage

import "sync"

type fileStorageConfig interface {
	StoreFilePath() string
}

type fileStorage struct {
	storeFile string
	lock      sync.RWMutex
}

func NewFileStorage(config fileStorageConfig) MetricsStorage {

	return &fileStorage{
		storeFile: config.StoreFilePath(),
	}
}

func (f *fileStorage) AddGaugeMetricValue(name string, value float64) float64 {
	f.lock.Lock()
	defer f.lock.Unlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) AddCounterMetricValue(name string, value int64) int64 {
	f.lock.Lock()
	defer f.lock.Unlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) GetMetricValues() map[string]map[string]string {
	f.lock.RLock()
	defer f.lock.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) GetMetricValue(metricType string, metricName string) (float64, bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) Restore(rawMetrics string) {
	//TODO implement me
	panic("implement me")
}

func (s *fileStorage) Close() {
	//TODO implement me
	panic("implement me")
}
