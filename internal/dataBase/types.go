package dataBase

import (
	"database/sql/driver"
	"io"
)

type DataBase interface {
	driver.Pinger
	io.Closer

	//WriteRecords(records []DBRecord) error
}

type DBRecord struct {
	MetricType string
	Name       string
	Value      float64
}
