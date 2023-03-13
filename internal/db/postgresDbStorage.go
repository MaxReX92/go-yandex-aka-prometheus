package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	_ "github.com/jackc/pgx/v5/stdlib"
	"io"
)

type DbStorage interface {
	driver.Pinger
	io.Closer
}

type PostgresDbStorageConfig interface {
	GetConnectionString() string
}

type PostgresDbStorage struct {
	conn *sql.DB
}

func NewPostgresDbStorage(conf PostgresDbStorageConfig) (*PostgresDbStorage, error) {
	connection, err := sql.Open("pgx", "host=localhost user=Max database=postgres password=1234")
	if err != nil {
		return nil, err
	}

	return &PostgresDbStorage{conn: connection}, nil
}

func (p *PostgresDbStorage) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *PostgresDbStorage) Close() error {
	return p.conn.Close()
}
