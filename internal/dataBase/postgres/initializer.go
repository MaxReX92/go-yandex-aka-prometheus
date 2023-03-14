package postgres

import (
	"context"
	"database/sql"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const createMetricTypeTableCommand = "CREATE TABLE IF NOT EXISTS metricType (" +
	"id SMALLSERIAL PRIMARY KEY," +
	"name CHAR(100));"

const createMetricTableCommand = "CREATE TABLE IF NOT EXISTS metric (" +
	"id SERIAL PRIMARY KEY," +
	"name CHAR(100)," +
	"typeId SMALLSERIAL," +
	"value DOUBLE PRECISION);"

const createMetricIndexCommand = "CREATE UNIQUE INDEX IF NOT EXISTS metric_name_type_idx " +
	"ON metric (name, typeId);"

func initDb(ctx context.Context, connectionString string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.ErrorFormat("Fail to open db connection: %v", err)
		return nil, err
	}

	err = conn.PingContext(ctx)
	if err != nil {
		logger.ErrorFormat("Fail to ping db connection: %v", err)
		return nil, err
	}

	_, err = conn.ExecContext(ctx, createMetricTypeTableCommand)
	if err != nil {
		logger.ErrorFormat("Fail to create metric type table: %v", err)
		return nil, err
	}

	_, err = conn.ExecContext(ctx, createMetricTableCommand)
	if err != nil {
		logger.ErrorFormat("Fail to create metric table: %v", err)
		return nil, err
	}

	_, err = conn.ExecContext(ctx, createMetricIndexCommand)
	if err != nil {
		logger.ErrorFormat("Fail to create metric index: %v", err)
		return nil, err
	}

	return conn, nil
}
