package storage

import (
	"bufio"
	"encoding/json"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"os"
	"sync"
)

type backupRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type fileStorageConfig interface {
	StoreFilePath() string
}

type fileStorage struct {
	filePath string
	lock     sync.RWMutex
}

func NewFileStorage(config fileStorageConfig) (MetricsStorage, error) {
	return &fileStorage{
		filePath: config.StoreFilePath(),
	}, nil
}

func (f *fileStorage) AddGaugeMetricValue(name string, value float64) (float64, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.updateMetric("gauge", name, parser.FloatToString(value))
}

func (f *fileStorage) AddCounterMetricValue(name string, value int64) (int64, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

}

func (f *fileStorage) GetMetricValue(metricType string, metricName string) (float64, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) GetMetricValues() (map[string]map[string]string, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) Restore(metricValues map[string]map[string]string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.fileStream == nil {
		return nil
	}

	fileName := f.fileStream.Name()
	err := f.fileStream.Close()
	if err != nil {
		return err
	}

	fileStream, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 644)
	if err != nil {
		return err
	}
	f.fileStream = fileStream
	encoder := json.NewEncoder(fileStream)

	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			err = appendToEnd(encoder, metricType, metricName, metricValue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *fileStorage) updateMetric(metricType string, metricName string, stringValue string) error {
	if f.filePath == "" {
		return nil
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	fileStream, err := os.OpenFile(f.filePath, os.O_RDWR|os.O_CREATE)
	if err != nil {
		return err
	}
	defer func(fileStream *os.File) {
		err := fileStream.Close()
		if err != nil {
			logger.ErrorFormat("Fail to close file: %v", err.Error())
		}
	}(fileStream)

	scanner := bufio.NewScanner(fileStream)
	for scanner.Scan() {

		scanner.Bytes()
	}
}

func appendToEnd(encoder *json.Encoder, metricType string, metricName string, stringValue string) error {
	return encoder.Encode(&backupRecord{
		Type:  metricType,
		Name:  metricName,
		Value: stringValue,
	})
}

func (f *fileStorage) Close() {
	if f.fileStream != nil {
		err := f.fileStream.Close()
		if err != nil {
			logger.ErrorFormat("Fail to close backup file: %v", err.Error())
		}
	}
}
