package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
)

// DataBase is a main abstraction for work with metric records.
type DataBase interface {
	driver.Pinger
	io.Closer

	// UpdateRecords update record values in database.
	UpdateRecords(ctx context.Context, records []*DBRecord) error

	// ReadRecord return metric db record from database.
	ReadRecord(ctx context.Context, metricType string, metricName string) (*DBRecord, error)

	// ReadAll return all metric db records from database.
	ReadAll(ctx context.Context) ([]*DBRecord, error)
}

// DBRecord represent metric in data base model.
type DBRecord struct {
	MetricType sql.NullString
	Name       sql.NullString
	Value      sql.NullFloat64
}
