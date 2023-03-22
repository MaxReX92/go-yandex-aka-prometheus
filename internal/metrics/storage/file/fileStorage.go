package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type storageRecord struct {
	Type  string `json:"types"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type storageRecords []*storageRecord

type fileStorageConfig interface {
	StoreFilePath() string
}

type fileStorage struct {
	filePath string
	lock     sync.Mutex
}

func NewFileStorage(config fileStorageConfig) storage.MetricsStorage {
	result := &fileStorage{
		filePath: config.StoreFilePath(),
	}

	if _, err := os.Stat(result.filePath); err != nil && result.filePath != "" && errors.Is(err, os.ErrNotExist) {
		logger.InfoFormat("Init storage file in %v", result.filePath)
		err = result.writeRecordsToFile(storageRecords{})
		if err != nil {
			logger.ErrorFormat("failed to init storage file: %v", err)
		}
	}

	return result
}

func (f *fileStorage) AddMetricValues(ctx context.Context, metricsList []metrics.Metric) ([]metrics.Metric, error) {
	return metricsList, f.updateMetrics(metricsList)
}

func (f *fileStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	records, err := f.readRecordsFromFile(func(record *storageRecord) bool {
		return record.Type == metricType && record.Name == metricName
	})
	if err != nil {
		return nil, logger.WrapError("read records from file", err)
	}
	if len(records) != 1 {
		return nil, logger.WrapError(fmt.Sprintf("get metric with name %v and type %v ", metricName, metricType), metrics.ErrMetricNotFound)
	}

	return f.toMetric(*records[0])
}

func (f *fileStorage) GetMetricValues(context.Context) (map[string]map[string]string, error) {
	records, err := f.readRecordsFromFile(func(record *storageRecord) bool { return true })
	if err != nil {
		return nil, logger.WrapError("read records from file", err)
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

	return result, nil
}

func (f *fileStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	var records storageRecords
	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			records = append(records, &storageRecord{
				Type:  metricType,
				Name:  metricName,
				Value: metricValue,
			})
		}
	}

	return f.writeRecordsToFile(records)
}

func (f *fileStorage) updateMetrics(metricsList []metrics.Metric) error {
	// Read and write
	return f.workWithFile(os.O_CREATE|os.O_RDWR, func(fileStream *os.File) error {
		metricsMap := map[string]metrics.Metric{} // contains?
		for _, metric := range metricsList {
			metricsMap[metric.GetType()+metric.GetName()] = metric
		}

		records, err := f.readRecords(fileStream, func(record *storageRecord) bool {
			_, found := metricsMap[record.Type+record.Name]
			return !found
		})
		if err != nil {
			return logger.WrapError("read records", err)
		}

		_, err = fileStream.Seek(0, io.SeekStart)
		if err != nil {
			return logger.WrapError("seek pointer", err)
		}
		err = fileStream.Truncate(0)
		if err != nil {
			return logger.WrapError("truncate file stream", err)
		}

		for _, metric := range metricsList {
			records = append(records, &storageRecord{
				Type:  metric.GetType(),
				Name:  metric.GetName(),
				Value: metric.GetStringValue(),
			})
		}

		return f.writeRecords(fileStream, records)
	})
}

func (f *fileStorage) readRecordsFromFile(isValid func(*storageRecord) bool) (storageRecords, error) {
	// ReadOnly
	return f.workWithFileResult(os.O_CREATE|os.O_RDONLY, func(fileStream *os.File) (storageRecords, error) {
		return f.readRecords(fileStream, isValid)
	})
}

func (f *fileStorage) readRecords(fileStream *os.File, isValid func(*storageRecord) bool) (storageRecords, error) {
	var records storageRecords
	err := json.NewDecoder(fileStream).Decode(&records)
	if err != nil {
		return nil, logger.WrapError("decode storage", err)
	}

	result := storageRecords{}
	for _, record := range records {
		if isValid(record) {
			result = append(result, record)
		}
	}

	return result, nil
}

func (f *fileStorage) writeRecordsToFile(records storageRecords) error {
	// WriteOnly
	return f.workWithFile(os.O_CREATE|os.O_WRONLY, func(fileStream *os.File) error {
		return f.writeRecords(fileStream, records)
	})
}

func (f *fileStorage) writeRecords(fileStream *os.File, records storageRecords) error {
	encoder := json.NewEncoder(fileStream)
	encoder.SetIndent("", " ")
	return encoder.Encode(records)
}

func (f *fileStorage) workWithFile(flag int, work func(file *os.File) error) error {
	_, err := f.workWithFileResult(flag, func(fileStream *os.File) (storageRecords, error) {
		return nil, work(fileStream)
	})
	return err
}

func (f *fileStorage) workWithFileResult(flag int, work func(file *os.File) (storageRecords, error)) (storageRecords, error) {
	if f.filePath == "" {
		return nil, nil
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	fileStream, err := os.OpenFile(f.filePath, flag, 0644)
	if err != nil {
		return nil, logger.WrapError("open file", err)
	}
	defer func(fileStream *os.File) {
		err = fileStream.Close()
		if err != nil {
			logger.ErrorFormat("failed to close file: %v", err)
		}
	}(fileStream)

	return work(fileStream)
}

func (f *fileStorage) toMetric(record storageRecord) (metrics.Metric, error) {
	var metric metrics.Metric
	switch record.Type {
	case "counter":
		metric = types.NewCounterMetric(record.Name)
	case "gauge":
		metric = types.NewGaugeMetric(record.Name)
	default:
		return nil, logger.WrapError(fmt.Sprintf("convert to metric with type %s", record.Type), metrics.ErrUnknownMetricType)
	}

	value, err := parser.ToFloat64(record.Value)
	if err != nil {
		return nil, logger.WrapError("parse record value", err)
	}

	metric.SetValue(value)
	return metric, nil
}
