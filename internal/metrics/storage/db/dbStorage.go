package db

import (
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/db"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
)

type dbStorage struct {
	dataBase db.DataBase
}

func NewDBStorage(dataBase db.DataBase) storage.MetricsStorage {
	return &dbStorage{dataBase: dataBase}
}

func (d dbStorage) AddMetricValue(metric metrics.Metric) (metrics.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (d dbStorage) GetMetricValues() (map[string]map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (d dbStorage) GetMetric(metricType string, metricName string) (metrics.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (d dbStorage) Restore(metricValues map[string]map[string]string) error {
	//TODO implement me
	panic("implement me")
}
