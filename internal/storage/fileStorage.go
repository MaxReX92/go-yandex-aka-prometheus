package storage

type fileStorageConfig interface {
	StoreFilePath() string
}

type fileStorage struct {
	storeFile string
}

func NewFileStorage(config fileStorageConfig) MetricsStorage {
	return &fileStorage{
		storeFile: config.StoreFilePath(),
	}
}

func (f *fileStorage) AddGaugeMetricValue(name string, value float64) float64 {
	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) AddCounterMetricValue(name string, value int64) int64 {
	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) GetMetricValues() map[string]map[string]string {
	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) GetMetricValue(metricType string, metricName string) (float64, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *fileStorage) Restore(rawMetrics string) {
	//TODO implement me
	panic("implement me")
}
