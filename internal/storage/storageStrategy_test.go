package storage

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type configMock struct {
	mock.Mock
}

type metricStorageMock struct {
	mock.Mock
}

const (
	metricType               = "metricType"
	metricName               = "metricName"
	metricValueInt   int64   = 100
	metricValueFloat float64 = 100
)

var (
	errTest = errors.New("errTest")
)

func TestStorageStrategy_AddGaugeMetricValue(t *testing.T) {

	tests := []struct {
		name                  string
		syncMode              bool
		inMemoryStorageResult float64
		inMemoryStorageError  error
		fileStorageResult     float64
		fileStorageErrorError error
		expectedResult        float64
		expectedError         error
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
			name:                  "noSync_fileStorage_error",
			syncMode:              false,
			inMemoryStorageResult: metricValueFloat,
			fileStorageErrorError: errTest,
			expectedResult:        metricValueFloat,
		},
		{
			name:                  "sync_fileStorage_error",
			syncMode:              true,
			inMemoryStorageResult: metricValueFloat,
			fileStorageErrorError: errTest,
			expectedError:         errTest,
		},
		{
			name:                  "noSync_success",
			syncMode:              false,
			inMemoryStorageResult: metricValueFloat,
			fileStorageResult:     metricValueFloat,
			expectedResult:        metricValueFloat,
		},
		{
			name:                  "sync_success",
			syncMode:              true,
			inMemoryStorageResult: metricValueFloat,
			fileStorageResult:     metricValueFloat,
			expectedResult:        metricValueFloat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddGaugeMetricValue", metricName, metricValueFloat).Return(tt.inMemoryStorageResult, tt.inMemoryStorageError)
			fileStorageMock.On("AddGaugeMetricValue", metricName, tt.inMemoryStorageResult).Return(tt.fileStorageResult, tt.fileStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualResult, actualError := strategy.AddGaugeMetricValue(metricName, metricValueFloat)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "AddGaugeMetricValue", metricName, metricValueFloat)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					fileStorageMock.AssertCalled(t, "AddGaugeMetricValue", metricName, tt.inMemoryStorageResult)
				} else {
					fileStorageMock.AssertNotCalled(t, "AddGaugeMetricValue", mock.Anything, mock.Anything)
				}
			} else {
				fileStorageMock.AssertNotCalled(t, "AddGaugeMetricValue", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_AddCounterMetricValue(t *testing.T) {

	tests := []struct {
		name                  string
		syncMode              bool
		inMemoryStorageResult int64
		inMemoryStorageError  error
		fileStorageResult     int64
		fileStorageErrorError error
		expectedResult        int64
		expectedError         error
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
			name:                  "noSync_fileStorage_error",
			syncMode:              false,
			inMemoryStorageResult: metricValueInt,
			fileStorageErrorError: errTest,
			expectedResult:        metricValueInt,
		},
		{
			name:                  "sync_fileStorage_error",
			syncMode:              true,
			inMemoryStorageResult: metricValueInt,
			fileStorageErrorError: errTest,
			expectedError:         errTest,
		},
		{
			name:                  "noSync_success",
			syncMode:              false,
			inMemoryStorageResult: metricValueInt,
			fileStorageResult:     metricValueInt,
			expectedResult:        metricValueInt,
		},
		{
			name:                  "sync_success",
			syncMode:              true,
			inMemoryStorageResult: metricValueInt,
			fileStorageResult:     metricValueInt,
			expectedResult:        metricValueInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddCounterMetricValue", metricName, metricValueInt).Return(tt.inMemoryStorageResult, tt.inMemoryStorageError)
			fileStorageMock.On("AddCounterMetricValue", metricName, tt.inMemoryStorageResult).Return(tt.fileStorageResult, tt.fileStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualResult, actualError := strategy.AddCounterMetricValue(metricName, metricValueInt)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "AddCounterMetricValue", metricName, metricValueInt)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					fileStorageMock.AssertCalled(t, "AddCounterMetricValue", metricName, tt.inMemoryStorageResult)
				} else {
					fileStorageMock.AssertNotCalled(t, "AddCounterMetricValue", mock.Anything, mock.Anything)
				}
			} else {
				fileStorageMock.AssertNotCalled(t, "AddCounterMetricValue", mock.Anything, mock.Anything)
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
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues").Return(tt.storageResult, tt.storageError)
			fileStorageMock.On("GetMetricValues").Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualResult, actualError := strategy.GetMetricValues()

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues")
			fileStorageMock.AssertNotCalled(t, "GetMetricValues")
		})
	}
}

func TestStorageStrategy_GetMetricValue(t *testing.T) {

	tests := []struct {
		name           string
		syncMode       bool
		storageResult  float64
		storageError   error
		expectedResult float64
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
			storageResult:  metricValueFloat,
			expectedResult: metricValueFloat,
		},
		{
			name:           "sync_success",
			syncMode:       true,
			storageResult:  metricValueFloat,
			expectedResult: metricValueFloat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValue", metricType, metricName).Return(tt.storageResult, tt.storageError)
			fileStorageMock.On("GetMetricValue", metricType, metricName).Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualResult, actualError := strategy.GetMetricValue(metricType, metricName)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValue", metricType, metricName)
			fileStorageMock.AssertNotCalled(t, "GetMetricValue", mock.Anything, mock.Anything)
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
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("Restore", values).Return(tt.storageError)
			fileStorageMock.On("Restore", values).Return(tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualError := strategy.Restore(values)

			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "Restore", values)
			fileStorageMock.AssertNotCalled(t, "Restore", mock.Anything)
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
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues").Return(tt.currentStateValues, tt.currentStateError)
			fileStorageMock.On("Restore", tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualError := strategy.CreateBackup()

			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues")

			if tt.currentStateError == nil {
				fileStorageMock.AssertCalled(t, "Restore", tt.currentStateValues)
			} else {
				fileStorageMock.AssertNotCalled(t, "Restore", mock.Anything)
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
			fileStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			fileStorageMock.On("GetMetricValues").Return(tt.currentStateValues, tt.currentStateError)
			inMemoryStorageMock.On("Restore", tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, fileStorageMock)
			actualError := strategy.RestoreFromBackup()

			assert.Equal(t, tt.expectedError, actualError)

			fileStorageMock.AssertCalled(t, "GetMetricValues")

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

func (s *metricStorageMock) AddGaugeMetricValue(name string, value float64) (float64, error) {
	args := s.Called(name, value)
	return args.Get(0).(float64), args.Error(1)
}

func (s *metricStorageMock) AddCounterMetricValue(name string, value int64) (int64, error) {
	args := s.Called(name, value)
	return args.Get(0).(int64), args.Error(1)
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
