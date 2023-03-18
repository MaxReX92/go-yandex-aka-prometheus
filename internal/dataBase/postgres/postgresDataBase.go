package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
)

type PostgresDataaBaseConfig interface {
	GetConnectionString() string
}

type postgresDataBase struct {
	conn *sql.DB
}

func (p *postgresDataBase) UpdateRecords(ctx context.Context, records []dataBase.DBRecord) error {
	const getCommand string = "" +
		"select m.id from metric m " +
		"join metricType mt on m.typeId = mt.id " +
		"where mt.name = @metricType and m.name = 'dsfsdf' "

	const incertCommand string = "" +
		"select m.type, m.name, m.value" +
		"from metric m" +
		"where m.type = @metricType and m.name = @metricName"

	const updateCommand string = "" +
		"select m.type, m.name, m.value" +
		"from metric m" +
		"where m.type = @metricType and m.name = @metricName"

	tx, err := p.conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return err
	}

	for _, record := range records {
		result := tx.QueryRowContext(ctx, getCommand, pgx.NamedArgs{
			"metricType": record.MetricType,
			"metricName": record.Name})

		var metricId uint
		err = result.Scan(&metricId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// insert
			} else {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.ErrorFormat("fail to rollback transaction: %v", rollbackErr)
				}
				return err
			}
		} else {
			//update
		}
	}

	return nil
}

func doInTransaction() {

}

func (p *postgresDataBase) ReadRecord(ctx context.Context, metricType string, metricName string) (dataBase.DBRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (p *postgresDataBase) ReadAll(context.Context) ([]dataBase.DBRecord, error) {
	//TODO implement me
	panic("implement me")
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
