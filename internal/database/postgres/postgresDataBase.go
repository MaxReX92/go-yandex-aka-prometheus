package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

// PostgresDataaBaseConfig contains required postgres connection settings.
type PostgresDataaBaseConfig interface {
	GetConnectionString() string
}

type postgresDataBase struct {
	conn *sql.DB
}

// NewPostgresDataBase create new instance of postgres db connector.
func NewPostgresDataBase(ctx context.Context, conf PostgresDataaBaseConfig) (*postgresDataBase, error) {
	conn, err := initDB(ctx, conf.GetConnectionString())
	if err != nil {
		return nil, logger.WrapError("init postgresql database", err)
	}

	return &postgresDataBase{conn: conn}, nil
}

func (p *postgresDataBase) UpdateRecords(ctx context.Context, records []*database.DBRecord) error {
	return p.callInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for _, record := range records {
			// statements for stored procedure are stored in a db
			_, err := tx.ExecContext(ctx, "CALL UpdateOrCreateMetric(@metricType, @metricName, @metricValue)", pgx.NamedArgs{
				"metricType":  record.MetricType.String,
				"metricName":  record.Name.String,
				"metricValue": record.Value.Float64,
			})
			if err != nil {
				return logger.WrapError("update records in postgresql database", err)
			}
		}

		return nil
	})
}

func (p *postgresDataBase) ReadRecord(ctx context.Context, metricType string, metricName string) (*database.DBRecord, error) {
	result, err := p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBRecord, error) {
		const command = "" +
			"SELECT mt.name, m.name, m.value " +
			"FROM metric m " +
			"JOIN metricType mt ON m.typeId = mt.id " +
			"WHERE " +
			"	m.name = @metricName " +
			"	and mt.name = @metricType"

		return p.readRecords(ctx, tx, command, pgx.NamedArgs{
			"metricType": metricType,
			"metricName": metricName,
		})
	})
	if err != nil {
		return nil, logger.WrapError("read records from postgresql database", err)
	}

	count := len(result)
	if count == 0 {
		return nil, nil
	}

	if count > 1 {
		logger.ErrorFormat("More than one metric in logical primary key: %v, %v", metricType, metricName)
	}

	return result[0], nil
}

func (p *postgresDataBase) ReadAll(ctx context.Context) ([]*database.DBRecord, error) {
	return p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBRecord, error) {
		const command = "" +
			"SELECT mt.name, m.name, m.value " +
			"FROM metric m " +
			"JOIN metricType mt on m.typeId = mt.id"

		return p.readRecords(ctx, tx, command)
	})
}

func (p *postgresDataBase) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *postgresDataBase) Close() error {
	return p.conn.Close()
}

func (p *postgresDataBase) callInTransaction(ctx context.Context, action func(context.Context, *sql.Tx) error) error {
	_, err := p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBRecord, error) {
		return nil, action(ctx, tx)
	})

	return err
}

func (p *postgresDataBase) callInTransactionResult(ctx context.Context, action func(context.Context, *sql.Tx) ([]*database.DBRecord, error)) ([]*database.DBRecord, error) {
	tx, err := p.conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return nil, logger.WrapError("begin transaction in postgresql database", err)
	}

	result, err := action(ctx, tx)
	if err != nil {
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			logger.ErrorFormat("failed to rollback transaction: %v", rollbackError)
		}

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, logger.WrapError("commit transaction", err)
	}

	return result, nil
}

func (p *postgresDataBase) readRecords(ctx context.Context, tx *sql.Tx, command string, args ...any) ([]*database.DBRecord, error) {
	rows, err := tx.QueryContext(ctx, command, args...)
	if err != nil {
		return nil, logger.WrapError("call query", err)
	}
	defer rows.Close()

	result := []*database.DBRecord{}
	for rows.Next() {
		var record database.DBRecord
		err = rows.Scan(&record.MetricType, &record.Name, &record.Value)
		if err != nil {
			return nil, logger.WrapError("scan rows", err)
		}

		result = append(result, &record)
	}

	err = rows.Err()
	if err != nil {
		return nil, logger.WrapError("get rows", err)
	}

	return result, nil
}
