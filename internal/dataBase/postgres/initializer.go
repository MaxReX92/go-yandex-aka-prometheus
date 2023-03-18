package postgres

import (
	"context"
	"database/sql"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const createMetricTypeTableCommand = "" +
	"CREATE TABLE IF NOT EXISTS metricType ( " +
	"id SMALLSERIAL PRIMARY KEY, " +
	"name CHAR(100));"

const createMetricTableCommand = "" +
	"CREATE TABLE IF NOT EXISTS metric ( " +
	"id SERIAL PRIMARY KEY, " +
	"name CHAR(1000), " +
	"typeId SMALLSERIAL, " +
	"value DOUBLE PRECISION);"

const createMetricIndexCommand = "" +
	"CREATE UNIQUE INDEX IF NOT EXISTS metric_name_type_idx " +
	"ON metric (name, typeId);"

const createMetricTypeIdProcedureCommand = "" +
	"CREATE OR REPLACE PROCEDURE GetOrCreateMetricTypeId(typeName IN CHAR(100), typeId OUT SMALLINT) " +
	"LANGUAGE plpgsql " +
	"AS $$ " +
	"BEGIN " +
	"	typeId := (SELECT id FROM metricType WHERE name = typeName); " +
	"	IF typeId IS null THEN " +
	"		BEGIN " +
	"			INSERT INTO metricType(name) VALUES (typeName); " +
	"			typeId := (SELECT currval(pg_get_serial_sequence('metricType','id'))); " +
	"		END; " +
	"	END IF; " +
	"END;$$"

const createMetricIdProcedureCommand = "" +
	"CREATE OR REPLACE PROCEDURE GetOrCreateMetricId(metricTypeName IN CHAR(100), metricName IN CHAR(1000), metricId OUT INT) " +
	"LANGUAGE plpgsql " +
	"AS $$ " +
	"DECLARE " +
	"	metricTypeId smallint; " +
	"BEGIN " +
	"	CALL GetOrCreateMetricTypeId(metricTypeName, metricTypeId); " +
	"	metricId := (SELECT id FROM metric WHERE name = metricName AND typeId = metricTypeId); " +
	"	IF metricId IS null THEN " +
	"		BEGIN " +
	"			INSERT INTO metric(name, typeId) VALUES (metricName, metricTypeId); " +
	"			metricId := (SELECT currval(pg_get_serial_sequence('metric','id'))); " +
	"		END; " +
	"	END IF; " +
	"END;$$"

const createMetricProcedureCommand = "" +
	"CREATE OR REPLACE PROCEDURE UpdateOrCreateMetric(metricTypeName IN CHAR(100), metricName IN CHAR(1000), metricValue IN double precision) " +
	"LANGUAGE plpgsql " +
	"AS $$ " +
	"DECLARE " +
	"	metricId int; " +
	"BEGIN " +
	"	CALL GetOrCreateMetricId(metricTypeName, metricName, metricId); " +
	"	UPDATE metric SET value = metricValue WHERE id = metricId; " +
	"END;$$"

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
