package db

import (
	"context"
	"database/sql"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type dbStorage struct {
	dataBase database.DataBase
}

func NewDBStorage(dataBase database.DataBase) storage.MetricsStorage {
	return &dbStorage{dataBase: dataBase}
}

func (d *dbStorage) AddMetricValues(ctx context.Context, metricsList []metrics.Metric) ([]metrics.Metric, error) {
	dbRecords := make([]*database.DBRecord, len(metricsList))
	for i, metric := range metricsList {
		dbRecords[i] = toDBRecord(metric)
	}

	err := d.dataBase.UpdateRecords(ctx, dbRecords)
	if err != nil {
		return nil, logger.WrapError("update db record", err)
	}

	return metricsList, nil
}

func (d *dbStorage) GetMetricValues(ctx context.Context) (map[string]map[string]string, error) {
	records, err := d.dataBase.ReadAll(ctx)
	if err != nil {
		return nil, logger.WrapError("read all db records", err)
	}

	result := map[string]map[string]string{}
	for _, record := range records {
		if !record.MetricType.Valid {
			return nil, NewErrInvalidRecord("invalid record metric type")
		}

		metricType := record.MetricType.String
		metricsByType, ok := result[metricType]
		if !ok {
			metricsByType = map[string]string{}
			result[metricType] = metricsByType
		}

		if !record.Name.Valid {
			return nil, NewErrInvalidRecord("invalid record metric name")
		}
		metricName := record.Name.String

		if !record.Value.Valid {
			return nil, NewErrInvalidRecord("invalid record metric value")
		}

		metricsByType[metricName] = parser.FloatToString(record.Value.Float64)
	}

	return result, nil
}

func (d *dbStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	result, err := d.dataBase.ReadRecord(ctx, metricType, metricName)
	if err != nil {
		return nil, logger.WrapError("read db record", err)
	}

	return fromDBRecord(result)
}

func (d *dbStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	records := []*database.DBRecord{}
	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			value, err := parser.ToFloat64(metricValue)
			if err != nil {
				return logger.WrapError("parse metric value", err)
			}

			records = append(records, &database.DBRecord{
				MetricType: sql.NullString{String: metricType, Valid: true},
				Name:       sql.NullString{String: metricName, Valid: true},
				Value:      sql.NullFloat64{Float64: value, Valid: true},
			})
		}
	}

	return d.dataBase.UpdateRecords(ctx, records)
}
