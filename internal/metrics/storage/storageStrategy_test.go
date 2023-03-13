package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type configMock struct {
	mock.Mock
}

type metricStorageMock struct {
	mock.Mock
}

const (
	metricType          = "metricType"
	metricName          = "metricName"
	metricValue float64 = 100
)

var (
	errTest = errors.New("errTest")
)

func TestStorageStrategy_AddGaugeMetricValue(t *testing.T) {

	tests := []struct {
		name                    string
		syncMode                bool
		inMemoryStorageError    error
		backupStorageErrorError error
		expectedResult          metrics.Metric
		expectedError           error
	}{
		{
			name:                 "noSync_inMemoryStorage_error",
			syncMode:             false,
			inMemoryStorageError: errTest,
			expectedError:        errTest,
		},
		{
			name:                 "sync_inMemoryStorage_error",
			syncMode:             true,
			inMemoryStorageError: errTest,
			expectedError:        errTest,
		},
		{
			name:                    "noSync_backupStorage_error",
			syncMode:                false,
			backupStorageErrorError: errTest,
			expectedResult:          test.CreateGaugeMetric("resultMetric", 100),
		},
		{
			name:                    "sync_backupStorage_error",
			syncMode:                true,
			backupStorageErrorError: errTest,
			expectedError:           errTest,
		},
		{
			name:           "noSync_success",
			syncMode:       false,
			expectedResult: test.CreateGaugeMetric("resultMetric", 100),
		},
		{
			name:           "sync_success",
			syncMode:       true,
			expectedResult: test.CreateGaugeMetric("resultMetric", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			gaugeMetric := test.CreateGaugeMetric(metricName, metricValue)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddMetricValue", gaugeMetric).Return(tt.expectedResult, tt.inMemoryStorageError)
			backupStorageMock.On("AddMetricValue", tt.expectedResult).Return(tt.expectedResult, tt.backupStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.AddMetricValue(gaugeMetric)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "AddMetricValue", gaugeMetric)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					backupStorageMock.AssertCalled(t, "AddMetricValue", tt.expectedResult)
				} else {
					backupStorageMock.AssertNotCalled(t, "AddMetricValue", mock.Anything)
				}
			} else {
				backupStorageMock.AssertNotCalled(t, "AddMetricValue", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_AddCounterMetricValue(t *testing.T) {

	tests := []struct {
		name                    string
		syncMode                bool
		inMemoryStorageError    error
		backupStorageErrorError error
		expectedResult          metrics.Metric
		expectedError           error
	}{
		{
			name:                 "noSync_inMemoryStorage_error",
			syncMode:             false,
			inMemoryStorageError: errTest,
			expectedError:        errTest,
		},
		{
			name:                 "sync_inMemoryStorage_error",
			syncMode:             true,
			inMemoryStorageError: errTest,
			expectedError:        errTest,
		},
		{
			name:                    "noSync_backupStorage_error",
			syncMode:                false,
			backupStorageErrorError: errTest,
			expectedResult:          test.CreateCounterMetric("resultMetric", 100),
		},
		{
			name:                    "sync_backupStorage_error",
			syncMode:                true,
			backupStorageErrorError: errTest,
			expectedError:           errTest,
		},
		{
			name:           "noSync_success",
			syncMode:       false,
			expectedResult: test.CreateCounterMetric("resultMetric", 100),
		},
		{
			name:           "sync_success",
			syncMode:       true,
			expectedResult: test.CreateCounterMetric("resultMetric", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			counterMetric := test.CreateCounterMetric(metricName, metricValue)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddMetricValue", counterMetric).Return(tt.expectedResult, tt.inMemoryStorageError)
			backupStorageMock.On("AddMetricValue", tt.expectedResult).Return(tt.expectedResult, tt.backupStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.AddMetricValue(counterMetric)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "AddMetricValue", counterMetric)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					backupStorageMock.AssertCalled(t, "AddMetricValue", tt.expectedResult)
				} else {
					backupStorageMock.AssertNotCalled(t, "AddMetricValue", mock.Anything)
				}
			} else {
				backupStorageMock.AssertNotCalled(t, "AddMetricValue", mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_GetMetricValues(t *testing.T) {

	result := map[string]map[string]string{}

	tests := []struct {
		name           string
		syncMode       bool
		storageResult  map[string]map[string]string
		storageError   error
		expectedResult map[string]map[string]string
		expectedError  error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:           "noSync_success",
			syncMode:       true,
			storageResult:  result,
			expectedResult: result,
		},
		{
			name:           "sync_success",
			syncMode:       true,
			storageResult:  result,
			expectedResult: result,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues").Return(tt.storageResult, tt.storageError)
			backupStorageMock.On("GetMetricValues").Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.GetMetricValues()

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues")
			backupStorageMock.AssertNotCalled(t, "GetMetricValues")
		})
	}
}

func TestStorageStrategy_GetMetric(t *testing.T) {

	resultMetric := test.CreateGaugeMetric(metricName, metricValue)
	tests := []struct {
		name           string
		syncMode       bool
		storageResult  metrics.Metric
		storageError   error
		expectedResult metrics.Metric
		expectedError  error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:           "noSync_success",
			syncMode:       true,
			storageResult:  resultMetric,
			expectedResult: resultMetric,
		},
		{
			name:           "sync_success",
			syncMode:       true,
			storageResult:  resultMetric,
			expectedResult: resultMetric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetric", metricType, metricName).Return(tt.storageResult, tt.storageError)
			backupStorageMock.On("GetMetric", metricType, metricName).Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.GetMetric(metricType, metricName)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetric", metricType, metricName)
			backupStorageMock.AssertNotCalled(t, "GetMetric", mock.Anything, mock.Anything)
		})
	}
}

func TestStorageStrategy_Restore(t *testing.T) {

	values := map[string]map[string]string{}

	tests := []struct {
		name          string
		syncMode      bool
		storageError  error
		expectedError error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  errTest,
			expectedError: errTest,
		},
		{
			name:     "noSync_success",
			syncMode: true,
		},
		{
			name:     "sync_success",
			syncMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("Restore", values).Return(tt.storageError)
			backupStorageMock.On("Restore", values).Return(tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.Restore(values)

			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "Restore", values)
			backupStorageMock.AssertNotCalled(t, "Restore", mock.Anything)
		})
	}
}

func TestStorageStrategy_CreateBackup(t *testing.T) {

	values := map[string]map[string]string{}

	tests := []struct {
		name               string
		syncMode           bool
		currentStateValues map[string]map[string]string
		currentStateError  error
		restoreError       error
		expectedError      error
	}{
		{
			name:              "noSync_currentState_error",
			syncMode:          false,
			currentStateError: errTest,
			expectedError:     errTest,
		},
		{
			name:              "sync_currentState_error",
			syncMode:          true,
			currentStateError: errTest,
			expectedError:     errTest,
		},
		{
			name:               "noSync_restore_error",
			syncMode:           false,
			currentStateValues: values,
			restoreError:       errTest,
			expectedError:      errTest,
		},
		{
			name:               "sync_restore_error",
			syncMode:           true,
			currentStateValues: values,
			restoreError:       errTest,
			expectedError:      errTest,
		},
		{
			name:               "noSync_success",
			syncMode:           true,
			currentStateValues: values,
		},
		{
			name:               "sync_success",
			syncMode:           true,
			currentStateValues: values,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues").Return(tt.currentStateValues, tt.currentStateError)
			backupStorageMock.On("Restore", tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.CreateBackup()

			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues")

			if tt.currentStateError == nil {
				backupStorageMock.AssertCalled(t, "Restore", tt.currentStateValues)
			} else {
				backupStorageMock.AssertNotCalled(t, "Restore", mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_RestoreFromBackup(t *testing.T) {

	values := map[string]map[string]string{}

	tests := []struct {
		name               string
		syncMode           bool
		currentStateValues map[string]map[string]string
		currentStateError  error
		restoreError       error
		expectedError      error
	}{
		{
			name:              "noSync_currentState_error",
			syncMode:          false,
			currentStateError: errTest,
			expectedError:     errTest,
		},
		{
			name:              "sync_currentState_error",
			syncMode:          true,
			currentStateError: errTest,
			expectedError:     errTest,
		},
		{
			name:               "noSync_restore_error",
			syncMode:           false,
			currentStateValues: values,
			restoreError:       errTest,
			expectedError:      errTest,
		},
		{
			name:               "sync_restore_error",
			syncMode:           true,
			currentStateValues: values,
			restoreError:       errTest,
			expectedError:      errTest,
		},
		{
			name:               "noSync_success",
			syncMode:           true,
			currentStateValues: values,
		},
		{
			name:               "sync_success",
			syncMode:           true,
			currentStateValues: values,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			backupStorageMock.On("GetMetricValues").Return(tt.currentStateValues, tt.currentStateError)
			inMemoryStorageMock.On("Restore", tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.RestoreFromBackup()

			assert.Equal(t, tt.expectedError, actualError)

			backupStorageMock.AssertCalled(t, "GetMetricValues")

			if tt.currentStateError == nil {
				inMemoryStorageMock.AssertCalled(t, "Restore", tt.currentStateValues)
			} else {
				inMemoryStorageMock.AssertNotCalled(t, "Restore", mock.Anything)
			}
		})
	}
}

func (c *configMock) SyncMode() bool {
	args := c.Called()
	return args.Bool(0)
}

func (s *metricStorageMock) GetMetric(metricType string, metricName string) (metrics.Metric, error) {
	args := s.Called(metricType, metricName)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	} else {
		return result.(metrics.Metric), args.Error(1)
	}
}

func (s *metricStorageMock) AddMetricValue(metric metrics.Metric) (metrics.Metric, error) {
	args := s.Called(metric)

	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	} else {
		return result.(metrics.Metric), args.Error(1)
	}
}

func (s *metricStorageMock) GetMetricValues() (map[string]map[string]string, error) {
	args := s.Called()
	return args.Get(0).(map[string]map[string]string), args.Error(1)
}

func (s *metricStorageMock) GetMetricValue(metricType string, metricName string) (float64, error) {
	args := s.Called(metricType, metricName)
	return args.Get(0).(float64), args.Error(1)
}

func (s *metricStorageMock) Restore(metricValues map[string]map[string]string) error {
	args := s.Called(metricValues)
	return args.Error(0)
}
