package db

import (
	"database/sql"
	"fmt"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

func toDBRecord(metric metrics.Metric) *database.DBRecord {
	return &database.DBRecord{
		MetricType: sql.NullString{String: metric.GetType(), Valid: true},
		Name:       sql.NullString{String: metric.GetName(), Valid: true},
		Value:      sql.NullFloat64{Float64: metric.GetValue(), Valid: true},
	}
}

func fromDBRecord(record *database.DBRecord) (metrics.Metric, error) {
	if !record.MetricType.Valid {
		return nil, NewErrInvalidRecord("invalid record metric type")
	}
	metricType := record.MetricType.String

	if !record.Name.Valid {
		return nil, NewErrInvalidRecord("invalid record metric name")
	}
	metricName := record.Name.String

	if !record.Value.Valid {
		return nil, NewErrInvalidRecord("invalid record metric value")
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
