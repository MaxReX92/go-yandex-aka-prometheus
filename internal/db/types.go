package db

import (
	"database/sql/driver"
	"io"
)

type DataBase interface {
	driver.Pinger
	io.Closer
}

type DBRecord struct {
	metricType string
	name       string
	value      float64
}
