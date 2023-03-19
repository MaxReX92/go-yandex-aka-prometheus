package file

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type config struct {
	filePath string
}

func TestFileStorage_New(t *testing.T) {

	tests := []struct {
		name     string
		filePath string
	}{
		{
			name: "empty_path",
		},
		{
			name:     "success",
			filePath: os.TempDir() + "TestFileStorage_New",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewFileStorage(&config{filePath: tt.filePath})
			assert.NotNil(t, storage)

			if tt.filePath != "" {
				defer func(name string) {
					_ = os.Remove(name)
				}(tt.filePath)

				actualRecords := readRecords(t, tt.filePath)
				assert.Empty(t, actualRecords)
			}
		})
	}
}

func TestFileStorage_AddGaugeMetricValue(t *testing.T) {

	tests := []struct {
		name            string
		values          []test.KeyValue
		expecredRecords storageRecords
	}{
		{
			name: "one_value",
			values: []test.KeyValue{
				{Key: "testMetric", Value: 100.001},
			},
			expecredRecords: storageRecords{
				{Type: "gauge", Name: "testMetric", Value: "100.001"},
			},
		},
		{
			name: "many_values",
			values: []test.KeyValue{
				{Key: "testMetric1", Value: 100.001},
				{Key: "testMetric2", Value: 200.002},
				{Key: "testMetric3", Value: 300.003},
			},
			expecredRecords: storageRecords{
				{Type: "gauge", Name: "testMetric1", Value: "100.001"},
				{Type: "gauge", Name: "testMetric2", Value: "200.002"},
				{Type: "gauge", Name: "testMetric3", Value: "300.003"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := os.TempDir() + "TestFileStorage_AddGaugeMetricValue"
			defer func(name string) {
				_ = os.Remove(name)
			}(filePath)

			storage := NewFileStorage(&config{filePath: filePath})

			metricsList := make([]metrics.Metric, len(tt.values))
			for i, m := range tt.values {
				metricsList[i] = test.CreateGaugeMetric(m.Key, m.Value)
			}

			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			actualRecords := readRecords(t, filePath)
			assert.Equal(t, tt.expecredRecords, actualRecords)
		})
	}
}

func TestFileStorage_AddCounterMetricValue(t *testing.T) {
	tests := []struct {
		name            string
		values          []test.KeyValue
		expectedRecords storageRecords
	}{
		{
			name: "one_value",
			values: []test.KeyValue{
				{Key: "testMetric", Value: 100},
			},
			expectedRecords: storageRecords{
				{Type: "counter", Name: "testMetric", Value: "100"},
			},
		},
		{
			name: "many_values",
			values: []test.KeyValue{
				{Key: "testMetric1", Value: 100},
				{Key: "testMetric2", Value: 200},
				{Key: "testMetric3", Value: 300},
			},
			expectedRecords: storageRecords{
				{Type: "counter", Name: "testMetric1", Value: "100"},
				{Type: "counter", Name: "testMetric2", Value: "200"},
				{Type: "counter", Name: "testMetric3", Value: "300"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := os.TempDir() + "TestFileStorage_AddCounterMetricValue"
			defer func(name string) {
				_ = os.Remove(name)
			}(filePath)

			storage := NewFileStorage(&config{filePath: filePath})

			metricsList := make([]metrics.Metric, len(tt.values))
			for i, m := range tt.values {
				metricsList[i] = test.CreateCounterMetric(m.Key, m.Value)
			}

			_, err := storage.AddMetricValues(context.Background(), metricsList)
			assert.NoError(t, err)

			actualRecords := readRecords(t, filePath)
			assert.Equal(t, tt.expectedRecords, actualRecords)
		})
	}
}

func TestFileStorage_GetMetric(t *testing.T) {

	expectedMetricType := "gauge"
	expectedMetricName := "expectedMetricName"
	expectedValue := float64(300)

	tests := []struct {
		name          string
		stored        storageRecords
		expectedError error
	}{
		{
			name:          "empty_store",
			stored:        storageRecords{},
			expectedError: errors.New("metrics with name expectedMetricName and types gauge not found"),
		},
		{
			name: "notFound",
			stored: storageRecords{
				{Type: "counter", Name: "metricName", Value: "100"},
			},
			expectedError: errors.New("metrics with name expectedMetricName and types gauge not found"),
		},
		{
			name: "success",
			stored: storageRecords{
				{Type: "counter", Name: "metricName", Value: "100"},
				{Type: expectedMetricType, Name: expectedMetricName, Value: "300"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := os.TempDir() + "TestFileStorage_GetMetricValue"
			defer func(name string) {
				_ = os.Remove(name)
			}(filePath)
			writeRecords(t, filePath, tt.stored)

			storage := NewFileStorage(&config{filePath: filePath})
			actualValue, err := storage.GetMetric(context.Background(), expectedMetricType, expectedMetricName)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedError == nil {
				assert.Equal(t, expectedValue, actualValue.GetValue())
			}
		})
	}
}

func readRecords(t *testing.T, filePath string) storageRecords {
	_, err := os.Stat(filePath)
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	records := storageRecords{}
	err = json.Unmarshal(content, &records)
	assert.NoError(t, err)

	return records
}

func writeRecords(t *testing.T, filePath string, records storageRecords) {
	fileStream, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	assert.NoError(t, err)
	defer func(fileStream *os.File) {
		_ = fileStream.Close()
	}(fileStream)

	err = json.NewEncoder(fileStream).Encode(records)
	assert.NoError(t, err)
}

func (c *config) StoreFilePath() string {
	return c.filePath
}
