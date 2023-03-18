package db

import (
	"context"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
)

type dbStorage struct {
	dataBase dataBase.DataBase
}

func NewDBStorage(dataBase dataBase.DataBase) storage.MetricsStorage {
	return &dbStorage{dataBase: dataBase}
}

func (d *dbStorage) AddMetricValue(ctx context.Context, metric metrics.Metric) (metrics.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (d *dbStorage) GetMetricValues(context.Context) (map[string]map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *dbStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (d *dbStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	//TODO implement me
	panic("implement me")
}
