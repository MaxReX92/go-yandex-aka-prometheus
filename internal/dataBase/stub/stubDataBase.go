package stub

import (
	"context"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/dataBase"
)

type StubDataBase struct {
}

func (s *StubDataBase) UpdateRecords(ctx context.Context, records []dataBase.DBRecord) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadRecord(ctx context.Context, metricType string, metricName string) (*dataBase.DBRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadAll(ctx context.Context) ([]dataBase.DBRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubDataBase) Ping(ctx context.Context) error {
	return nil
}

func (s *StubDataBase) Close() error {
	return nil
}
