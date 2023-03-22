package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type databaseMock struct {
	mock.Mock
}

func TestDbStorage_AddMetricValues(t *testing.T) {
	metricsList := []metrics.Metric{
		test.CreateCounterMetric("counterMetricName", 100),
		test.CreateGaugeMetric("gaugeMetricName", 200),
	}

	records := []*database.DBRecord{
		{
			MetricType: sql.NullString{Valid: true, String: "counter"},
			Name:       sql.NullString{Valid: true, String: "counterMetricName"},
			Value:      sql.NullFloat64{Valid: true, Float64: 100},
		},
		{
			MetricType: sql.NullString{Valid: true, String: "gauge"},
			Name:       sql.NullString{Valid: true, String: "gaugeMetricName"},
			Value:      sql.NullFloat64{Valid: true, Float64: 200},
		},
	}

	tests := []struct {
		name          string
		updateError   error
		expectedError error
	}{
		{
			name:          "update_error",
			updateError:   test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			dbMock := new(databaseMock)
			dbMock.On("UpdateRecords", ctx, records).Return(tt.updateError)

			storage := NewDBStorage(dbMock)
			actualResult, actualError := storage.AddMetricValues(ctx, metricsList)

			if tt.expectedError == nil {
				assert.Equal(t, metricsList, actualResult)
			} else {
				assert.Empty(t, actualResult)
			}
			assert.ErrorIs(t, actualError, tt.expectedError)
			dbMock.AssertCalled(t, "UpdateRecords", ctx, records)
		})
	}
}

func TestDbStorage_GetMetricValues(t *testing.T) {
	tests := []struct {
		name            string
		allRecords      []*database.DBRecord
		getRecordsError error

		expectedResult       map[string]map[string]string
		expectedErrorMessage string
	}{
		{
			name:                 "get_records_error",
			getRecordsError:      test.ErrTest,
			expectedErrorMessage: test.ErrTest.Error(),
		},
		{
			name: "invalid_record_type",
			allRecords: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: false},
			}},
			expectedErrorMessage: "invalid record metric type",
		},
		{
			name: "invalid_record_name",
			allRecords: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: false},
			}},
			expectedErrorMessage: "invalid record metric name",
		},
		{
			name: "invalid_record_value",
			allRecords: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: true, String: "testMetricName"},
				Value:      sql.NullFloat64{Valid: false},
			}},
			expectedErrorMessage: "invalid record metric value",
		},
		{
			name: "success",
			allRecords: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: true, String: "testMetricName1"},
				Value:      sql.NullFloat64{Valid: true, Float64: 100},
			}, {
				MetricType: sql.NullString{Valid: true, String: "gauge"},
				Name:       sql.NullString{Valid: true, String: "testMetricName2"},
				Value:      sql.NullFloat64{Valid: true, Float64: 200},
			}},
			expectedResult: map[string]map[string]string{
				"counter": {"testMetricName1": "100"},
				"gauge":   {"testMetricName2": "200"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			dbMock := new(databaseMock)
			dbMock.On("ReadAll", ctx).Return(tt.allRecords, tt.getRecordsError)

			storage := NewDBStorage(dbMock)
			actualResult, actualError := storage.GetMetricValues(ctx)
			assert.Equal(t, tt.expectedResult, actualResult)
			if tt.expectedErrorMessage != "" {
				assert.ErrorContains(t, actualError, tt.expectedErrorMessage)
			}

			dbMock.AssertCalled(t, "ReadAll", ctx)
		})
	}
}

func TestDbStorage_GetMetric(t *testing.T) {
	metricType := "testMetricType"
	metricName := "testMetricName"
	metricValie := float64(100)

	tests := []struct {
		name            string
		dbRecord        *database.DBRecord
		readRecordError error

		expectedResult       metrics.Metric
		expectedErrorMessage string
	}{
		{
			name:                 "read_record_error",
			readRecordError:      test.ErrTest,
			expectedErrorMessage: test.ErrTest.Error(),
		},
		{
			name:                 "invalid_metric_type",
			dbRecord:             &database.DBRecord{MetricType: sql.NullString{Valid: false}},
			expectedErrorMessage: "invalid record metric type",
		},
		{
			name: "invalid_metric_name",
			dbRecord: &database.DBRecord{
				MetricType: sql.NullString{Valid: true, String: metricType},
				Name:       sql.NullString{Valid: false},
			},
			expectedErrorMessage: "invalid record metric name",
		},
		{
			name: "invalid_metric_value",
			dbRecord: &database.DBRecord{
				MetricType: sql.NullString{Valid: true, String: metricType},
				Name:       sql.NullString{Valid: true, String: metricName},
				Value:      sql.NullFloat64{Valid: false},
			},
			expectedErrorMessage: "invalid record metric value",
		},
		{
			name: "unknown_metric_type",
			dbRecord: &database.DBRecord{
				MetricType: sql.NullString{Valid: true, String: metricType},
				Name:       sql.NullString{Valid: true, String: metricName},
				Value:      sql.NullFloat64{Valid: true, Float64: metricValie},
			},
			expectedErrorMessage: "failed to read record with type 'testMetricType': unknown metric type",
		},
		{
			name: "success_counter_metric",
			dbRecord: &database.DBRecord{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: true, String: metricName},
				Value:      sql.NullFloat64{Valid: true, Float64: metricValie},
			},
			expectedResult: test.CreateCounterMetric(metricName, metricValie),
		},
		{
			name: "success_gauge_metric",
			dbRecord: &database.DBRecord{
				MetricType: sql.NullString{Valid: true, String: "gauge"},
				Name:       sql.NullString{Valid: true, String: metricName},
				Value:      sql.NullFloat64{Valid: true, Float64: metricValie},
			},
			expectedResult: test.CreateGaugeMetric(metricName, metricValie),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			dbMock := new(databaseMock)
			dbMock.On("ReadRecord", ctx, metricType, metricName).Return(tt.dbRecord, tt.readRecordError)

			storage := NewDBStorage(dbMock)
			actualResult, actualError := storage.GetMetric(ctx, metricType, metricName)

			assert.Equal(t, tt.expectedResult, actualResult)
			if tt.expectedErrorMessage != "" {
				assert.ErrorContains(t, actualError, tt.expectedErrorMessage)
			}

			dbMock.AssertCalled(t, "ReadRecord", ctx, metricType, metricName)
		})
	}
}

func TestDbStorage_Restore(t *testing.T) {
	tests := []struct {
		name                 string
		metricValues         map[string]map[string]string
		records              []*database.DBRecord
		updateError          error
		expectedErrorMessage string
	}{
		{
			name:                 "invalid_value",
			metricValues:         map[string]map[string]string{"counter": {"metricName": "invalid_value"}},
			expectedErrorMessage: "failed to parse metric value",
		},
		{
			name:         "update_error",
			metricValues: map[string]map[string]string{"counter": {"metricName": "100"}},
			records: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: true, String: "metricName"},
				Value:      sql.NullFloat64{Valid: true, Float64: 100},
			}},
			updateError:          test.ErrTest,
			expectedErrorMessage: "test error message",
		},
		{
			name: "success",
			metricValues: map[string]map[string]string{
				"counter": {"counterMetricName": "100"},
				"gauge":   {"gaugeMetricName": "200"},
			},
			records: []*database.DBRecord{{
				MetricType: sql.NullString{Valid: true, String: "counter"},
				Name:       sql.NullString{Valid: true, String: "counterMetricName"},
				Value:      sql.NullFloat64{Valid: true, Float64: 100},
			}, {
				MetricType: sql.NullString{Valid: true, String: "gauge"},
				Name:       sql.NullString{Valid: true, String: "gaugeMetricName"},
				Value:      sql.NullFloat64{Valid: true, Float64: 200},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			dbMock := new(databaseMock)
			dbMock.On("UpdateRecords", ctx, tt.records).Return(tt.updateError)

			storage := NewDBStorage(dbMock)
			actualError := storage.Restore(ctx, tt.metricValues)

			if tt.expectedErrorMessage != "" {
				assert.ErrorContains(t, actualError, tt.expectedErrorMessage)
			} else {
				dbMock.AssertCalled(t, "UpdateRecords", ctx, tt.records)
			}
		})
	}
}

func (d *databaseMock) Ping(ctx context.Context) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *databaseMock) Close() error {
	args := d.Called()
	return args.Error(0)
}

func (d *databaseMock) UpdateRecords(ctx context.Context, records []*database.DBRecord) error {
	args := d.Called(ctx, records)
	return args.Error(0)
}

func (d *databaseMock) ReadRecord(ctx context.Context, metricType string, metricName string) (*database.DBRecord, error) {
	args := d.Called(ctx, metricType, metricName)
	return args.Get(0).(*database.DBRecord), args.Error(1)
}

func (d *databaseMock) ReadAll(ctx context.Context) ([]*database.DBRecord, error) {
	args := d.Called(ctx)
	return args.Get(0).([]*database.DBRecord), args.Error(1)
}
