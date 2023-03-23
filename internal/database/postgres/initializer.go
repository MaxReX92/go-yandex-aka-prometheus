package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

var scripts = map[string]string{
	"1 - create metric type table command": "" +
		"CREATE TABLE IF NOT EXISTS metricType ( " +
		"	id SMALLSERIAL PRIMARY KEY, " +
		"	name TEXT" +
		");",

	"2 - create metric table command": "" +
		"CREATE TABLE IF NOT EXISTS metric ( " +
		"	id SERIAL PRIMARY KEY, " +
		"	name TEXT, " +
		"	typeId SMALLSERIAL, " +
		"	value DOUBLE PRECISION " +
		");",

	"3 - create metric index command": "" +
		"CREATE UNIQUE INDEX IF NOT EXISTS metric_name_type_idx " +
		"ON metric (name, typeId);",

	"4 - create metric type id procedure command": "" +
		"CREATE OR REPLACE PROCEDURE GetOrCreateMetricTypeId(typeName IN TEXT, typeId OUT SMALLINT) " +
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
		"END;$$",

	"5 - create metric id procedure command": "" +
		"CREATE OR REPLACE PROCEDURE GetOrCreateMetricId(metricTypeName IN TEXT, metricName IN TEXT, metricId OUT INT) " +
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
		"END;$$",

	"6 - create metric procedure command": "" +
		"CREATE OR REPLACE PROCEDURE UpdateOrCreateMetric(metricTypeName IN TEXT, metricName IN TEXT, metricValue IN double precision) " +
		"LANGUAGE plpgsql " +
		"AS $$ " +
		"DECLARE " +
		"	metricId int; " +
		"BEGIN " +
		"	CALL GetOrCreateMetricId(metricTypeName, metricName, metricId); " +
		"	UPDATE metric SET value = metricValue WHERE id = metricId; " +
		"END;$$",
}

func initDB(ctx context.Context, connectionString string) (*sql.DB, error) {
	logger.Info("Initialize database schema")

	conn, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, logger.WrapError("open db connection", err)
	}

	err = conn.PingContext(ctx)
	if err != nil {
		return nil, logger.WrapError("ping db connection", err)
	}

	i := 0
	commandNames := make([]string, len(scripts))
	for commandName := range scripts {
		commandNames[i] = commandName
		i++
	}
	sort.Strings(commandNames)

	for _, commandName := range commandNames {
		command := scripts[commandName]
		logger.InfoFormat("Invoke %s", commandName)
		_, err = conn.ExecContext(ctx, command)
		if err != nil {
			return nil, logger.WrapError(fmt.Sprintf("invoke %s", commandName), err)
		}
	}

	return conn, nil
}
