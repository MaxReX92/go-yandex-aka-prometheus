package postgres

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
)

type PostgresDataaBaseConfig interface {
	GetConnectionString() string
}

type postgresDataBase struct {
	conn *sql.DB
}

func NewPostgresDataBase(ctx context.Context, conf PostgresDataaBaseConfig) (dataBase.DataBase, error) {
	conn, err := initDb(ctx, conf.GetConnectionString())
	if err != nil {
		return nil, err
	}

	return &postgresDataBase{conn: conn}, nil
}

func (p *postgresDataBase) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *postgresDataBase) Close() error {
	return p.conn.Close()
}
