package db

import (
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
)

func toDbRecord(metric metrics.Metric) *dataBase.DBRecord {
	return &dataBase.DBRecord{
		MetricType: metric.GetType(),
		Name:       metric.GetName(),
		Value:      metric.GetValue(),
	}
}

func fromDbRecord(record *dataBase.DBRecord) (metrics.Metric, error) {
	metricType := record.MetricType

	var metric metrics.Metric
	switch metricType {
	case "gauge":
		metric = types.NewGaugeMetric(record.Name)
	case "counter":
		metric = types.NewCounterMetric(record.Name)
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	metric.SetValue(record.Value)
	return metric, nil
}
