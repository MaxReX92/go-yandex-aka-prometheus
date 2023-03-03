package storage

import (
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

type backupRecords []*backupRecord

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
	records, err := f.readRecordsFromFile(func(record *backupRecord) bool {
		return record.Type == metricType && record.Name == metricName
	})
	if err != nil {
		return 0, err
	}
	if len(records) != 1 {
		return 0, fmt.Errorf("metrics with name %v and type %v not found", metricName, metricType)
	}

	return parser.ToFloat64(records[0].Value)
}

func (f *fileStorage) GetMetricValues() (map[string]map[string]string, error) {
	records, err := f.readRecordsFromFile(func(record *backupRecord) bool { return true })
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
	var records backupRecords
	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			records = append(records, &backupRecord{
				Type:  metricType,
				Name:  metricName,
				Value: metricValue,
			})
		}
	}

	return f.writeRecordsToFile(records)
}

func (f *fileStorage) updateMetric(metricType string, metricName string, stringValue string) error {
	// Read and write
	return f.workWithFile(os.O_RDWR|os.O_CREATE, func(fileStream *os.File) error {
		records, err := f.readRecords(fileStream, func(record *backupRecord) bool {
			return record.Type != metricType || record.Name != metricType
		})
		if err != nil {
			return err
		}

		records = append(records, &backupRecord{
			Type:  metricType,
			Name:  metricName,
			Value: stringValue,
		})
		return f.writeRecords(fileStream, records)
	})
}

func (f *fileStorage) readRecordsFromFile(isValid func(*backupRecord) bool) (backupRecords, error) {
	// ReadOnly
	return f.workWithFileResult(os.O_RDONLY|os.O_CREATE, func(fileStream *os.File) (backupRecords, error) {
		return f.readRecords(fileStream, isValid)
	})
}

func (f *fileStorage) readRecords(fileStream *os.File, isValid func(*backupRecord) bool) (backupRecords, error) {
	var records backupRecords
	err := json.NewDecoder(fileStream).Decode(&records)
	if err != nil {
		logger.ErrorFormat("Fail to decode backup: %v", err)
		return nil, err
	}

	result := backupRecords{}
	for _, record := range records {
		if isValid(record) {
			result = append(result, record)
		}
	}

	return result, nil
}

func (f *fileStorage) writeRecordsToFile(records backupRecords) error {
	// WriteOnly
	return f.workWithFile(os.O_WRONLY|os.O_CREATE, func(fileStream *os.File) error {
		return f.writeRecords(fileStream, records)
	})
}

func (f *fileStorage) writeRecords(fileStream *os.File, records backupRecords) error {
	return json.NewEncoder(fileStream).Encode(records)
}

func (f *fileStorage) workWithFile(flag int, work func(file *os.File) error) error {
	_, err := f.workWithFileResult(flag, func(fileStream *os.File) (backupRecords, error) {
		return nil, work(fileStream)
	})
	return err
}

func (f *fileStorage) workWithFileResult(flag int, work func(file *os.File) (backupRecords, error)) (backupRecords, error) {
	if f.filePath == "" {
		return nil, nil
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	fileStream, err := os.OpenFile(f.filePath, flag, 644)
	if err != nil {
		logger.ErrorFormat("Fail to open file: %v", err.Error())
		return nil, err
	}
	defer func(fileStream *os.File) {
		err = fileStream.Close()
		if err != nil {
			logger.ErrorFormat("Fail to close file: %v", err.Error())
		}
	}(fileStream)

	return work(fileStream)
}
