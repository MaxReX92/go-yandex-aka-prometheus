package handler

import (
	"context"
	"fmt"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/html"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/server"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/storage"
)

type requestHandler struct {
	dbStorage   database.DataBase
	pageBuilder html.PageBuilder
	storage     storage.MetricsStorage
}

func NewHandler(dbStorage database.DataBase, pageBuilder html.PageBuilder, storage storage.MetricsStorage) *requestHandler {
	return &requestHandler{
		dbStorage:   dbStorage,
		pageBuilder: pageBuilder,
		storage:     storage,
	}
}

func (h *requestHandler) UpdateMetricValues(ctx context.Context, metricValues []metrics.Metric) ([]metrics.Metric, error) {
	resultMetrics, err := h.storage.AddMetricValues(ctx, metricValues)
	if err != nil {
		return nil, logger.WrapError("update metric", err)
	}

	return resultMetrics, nil
}

func (h *requestHandler) GetMetricValue(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	metric, err := h.storage.GetMetric(ctx, metricType, metricName)
	if err != nil {
		return nil, logger.WrapError(fmt.Sprintf("get metric with type '%s' and name '%s'", metricType, metricName),
			server.ErrMetricNotFound)
	}

	return metric, nil
}

func (h *requestHandler) GetReportPage(ctx context.Context) (string, error) {
	values, err := h.storage.GetMetricValues(ctx)
	if err != nil {
		return "", logger.WrapError("get metric values", err)
	}

	return h.pageBuilder.BuildMetricsPage(values), nil
}

func (h *requestHandler) Ping(ctx context.Context) error {
	err := h.dbStorage.Ping(ctx)
	if err != nil {
		return logger.WrapError("ping database", err)
	}

	return nil
}
