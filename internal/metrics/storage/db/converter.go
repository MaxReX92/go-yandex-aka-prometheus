package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

var ErrInvalidRecord = errors.New("Invalid db record")

func toDbRecord(metric metrics.Metric) *dataBase.DBRecord {
	return &dataBase.DBRecord{
		MetricType: sql.NullString{String: metric.GetType(), Valid: true},
		Name:       sql.NullString{String: metric.GetName(), Valid: true},
		Value:      sql.NullFloat64{Float64: metric.GetValue(), Valid: true},
	}
}

func fromDbRecord(record *dataBase.DBRecord) (metrics.Metric, error) {
	if !record.MetricType.Valid {
		return nil, ErrInvalidRecord
	}
	metricType := record.MetricType.String

	if !record.Name.Valid {
		return nil, ErrInvalidRecord
	}
	metricName := record.Name.String

	if !record.Value.Valid {
		return nil, ErrInvalidRecord
	}
	value := record.Value.Float64

	var metric metrics.Metric
	switch metricType {
	case "gauge":
		metric = types.NewGaugeMetric(metricName)
	case "counter":
		metric = types.NewCounterMetric(metricName)
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	metric.SetValue(value)
	return metric, nil
}
