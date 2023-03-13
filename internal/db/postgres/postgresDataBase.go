package postgres

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/db"
)

type PostgresDataaBaseConfig interface {
	GetConnectionString() string
}

type postgresDataBase struct {
	conn *sql.DB
}

func NewPostgresDataBase(conf PostgresDataaBaseConfig) (db.DataBase, error) {
	connection, err := sql.Open("pgx", conf.GetConnectionString())
	if err != nil {
		return nil, err
	}

	return &postgresDataBase{conn: connection}, nil
}

func (p *postgresDataBase) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *postgresDataBase) Close() error {
	return p.conn.Close()
}
