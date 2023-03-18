package dataBase

import (
	"context"
	"database/sql/driver"
	"io"
)

type DataBase interface {
	driver.Pinger
	io.Closer

	UpdateRecords(ctx context.Context, records []DBRecord) error
	ReadRecord(ctx context.Context, metricType string, metricName string) (DBRecord, error)
	ReadAll(ctx context.Context) ([]DBRecord, error)
}

type DBRecord struct {
	MetricType string
	Name       string
	Value      float64
}
