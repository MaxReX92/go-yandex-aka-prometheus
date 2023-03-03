package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	lock     *sync.Mutex
}

func NewFileStorage(config fileStorageConfig) MetricsStorage {
	return &fileStorage{
		filePath: config.StoreFilePath(),
	}
}

func (f *fileStorage) AddGaugeMetricValue(name string, value float64) (float64, error) {
	return value, f.updateMetric("gauge", name, parser.FloatToString(value))
}

func (f *fileStorage) AddCounterMetricValue(name string, value int64) (int64, error) {
	return value, f.updateMetric("counter", name, parser.IntToString(value))
}

func (f *fileStorage) GetMetricValue(metricType string, metricName string) (float64, error) {
	record, err := workWithFileResult[*backupRecord](f.lock, f.filePath, os.O_RDONLY|os.O_CREATE, func(fileStream *os.File) (*backupRecord, error) {
		records, err := f.readRecords(fileStream, func(record *backupRecord) (bool, error) {
			return record.Type == metricType && record.Name == metricName, nil
		})
		if err != nil {
			return nil, err
		}
		if len(records) != 1 {
			return nil, fmt.Errorf("metrics with name %v and type %v not found", metricName, metricType)
		}

		return records[0], nil
	})

	if err != nil {
		return 0, err
	}
	return parser.ToFloat64(record.Value)
}

func (f *fileStorage) GetMetricValues() (map[string]map[string]string, error) {
	records, err := workWithFileResult[[]*backupRecord](f.lock, f.filePath, os.O_RDONLY|os.O_CREATE,
		func(fileStream *os.File) ([]*backupRecord, error) {
			return f.readRecords(fileStream, func(record *backupRecord) (bool, error) { return true, nil })
		})

	if err != nil {
		return nil, err
	}

	result := map[string]map[string]string{}
	for _, record := range records {
		metricsByType, ok := result[record.Type]
		if !ok {
			metricsByType = map[string]string{}
			result[record.Type] = metricsByType
		}

		metricsByType[record.Name] = record.Value
	}

	return result, err
}

func (f *fileStorage) Restore(metricValues map[string]map[string]string) error {
	return workWithFile(f.lock, f.filePath, os.O_WRONLY|os.O_CREATE, func(fileStream *os.File) error {
		records := []*backupRecord{}
		for metricType, metricsByType := range metricValues {
			for metricName, metricValue := range metricsByType {

				records = append(records, &backupRecord{
					Type:  metricType,
					Name:  metricName,
					Value: metricValue,
				})
			}
		}

		return writeToFile(fileStream, records)
	})
}

func (f *fileStorage) updateMetric(metricType string, metricName string, stringValue string) error {
	return workWithFile(f.lock, f.filePath, os.O_RDWR|os.O_CREATE, func(fileStream *os.File) error {
		records, err := f.readRecords(fileStream, func(record *backupRecord) (bool, error) {
			// skip for append to end
			return record.Type != metricType || record.Name != metricType, nil
		})

		if err != nil {
			return err
		}

		records = append(records, &backupRecord{
			Type:  metricType,
			Name:  metricName,
			Value: stringValue,
		})
		return writeToFile(fileStream, records)
	})
}

func (f *fileStorage) readRecords(fileStream *os.File, isValid func(*backupRecord) (bool, error)) ([]*backupRecord, error) {
	records := []*backupRecord{}
	scanner := bufio.NewScanner(fileStream)

	for scanner.Scan() {
		record := &backupRecord{}
		err := json.Unmarshal(scanner.Bytes(), record)
		if err != nil {
			logger.ErrorFormat("Unexpected string in backup file: %v", scanner.Text())
			continue
		}

		valid, err := isValid(record)
		if err != nil {
			return nil, err
		}
		if !valid {
			continue
		}

		records = append(records, record)
	}

	return records, scanner.Err()
}

func writeToFile(fileStream *os.File, records []*backupRecord) error {
	encoder := json.NewEncoder(fileStream)
	for _, record := range records {
		err := encoder.Encode(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func workWithFile(lock *sync.Mutex, filePath string, flag int, work func(file *os.File) error) error {
	_, err := workWithFileResult(lock, filePath, flag, func(fileStream *os.File) (int, error) {
		return 0, work(fileStream)
	})
	return err
}

func workWithFileResult[T any](lock *sync.Mutex, filePath string, flag int, work func(file *os.File) (T, error)) (T, error) {
	var defaultResult T
	if filePath == "" {
		return defaultResult, nil
	}

	lock.Lock()
	defer lock.Unlock()

	fileStream, err := os.OpenFile(filePath, flag, 644)
	if err != nil {
		return defaultResult, err
	}
	defer func(fileStream *os.File) {
		err = fileStream.Close()
		if err != nil {
			logger.ErrorFormat("Fail to close file: %v", err.Error())
		}
	}(fileStream)

	return work(fileStream)
}
