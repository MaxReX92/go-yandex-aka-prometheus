package db

import (
	"context"
	"database/sql"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
)

type dbStorage struct {
	dataBase dataBase.DataBase
}

func NewDBStorage(dataBase dataBase.DataBase) storage.MetricsStorage {
	return &dbStorage{dataBase: dataBase}
}

func (d *dbStorage) AddMetricValue(ctx context.Context, metric metrics.Metric) (metrics.Metric, error) {
	err := d.dataBase.UpdateRecords(ctx, []*dataBase.DBRecord{toDbRecord(metric)})
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func (d *dbStorage) GetMetricValues(ctx context.Context) (map[string]map[string]string, error) {
	records, err := d.dataBase.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string]map[string]string{}
	for _, record := range records {
		if !record.MetricType.Valid {
			return nil, ErrInvalidRecord
		}

		metricType := record.MetricType.String
		metricsByType, ok := result[metricType]
		if !ok {
			metricsByType = map[string]string{}
			result[metricType] = metricsByType
		}

		if !record.Name.Valid {
			return nil, ErrInvalidRecord
		}
		metricName := record.Name.String

		if !record.Value.Valid {
			return nil, ErrInvalidRecord
		}

		metricsByType[metricName] = parser.FloatToString(record.Value.Float64)
	}

	return result, nil
}

func (d *dbStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	result, err := d.dataBase.ReadRecord(ctx, metricType, metricName)
	if err != nil {
		return nil, err
	}

	return fromDbRecord(result)
}

func (d *dbStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {

	records := []*dataBase.DBRecord{}
	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			value, err := parser.ToFloat64(metricValue)
			if err != nil {
				return err
			}

			records = append(records, &dataBase.DBRecord{
				MetricType: sql.NullString{String: metricType, Valid: true},
				Name:       sql.NullString{String: metricName, Valid: true},
				Value:      sql.NullFloat64{Float64: value, Valid: true},
			})
		}
	}

	return d.dataBase.UpdateRecords(ctx, records)
}
