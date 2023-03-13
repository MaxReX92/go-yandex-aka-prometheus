package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	_ "github.com/jackc/pgx/v5/stdlib"
	"io"
)

type DBStorage interface {
	driver.Pinger
	io.Closer
}

type PostgresDBStorageConfig interface {
	GetConnectionString() string
}

type PostgresDBStorage struct {
	conn *sql.DB
}

func NewPostgresDBStorage(conf PostgresDBStorageConfig) (*PostgresDBStorage, error) {
	connection, err := sql.Open("pgx", conf.GetConnectionString())
	if err != nil {
		return nil, err
	}

	return &PostgresDBStorage{conn: connection}, nil
}

func (p *PostgresDBStorage) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *PostgresDBStorage) Close() error {
	return p.conn.Close()
}
