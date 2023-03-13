package dataBase

import (
	"database/sql/driver"
	"io"
)

type DataBase interface {
	driver.Pinger
	io.Closer
}

type DBRecord struct {
	MetricType string
	Name       string
	Value      float64
}
