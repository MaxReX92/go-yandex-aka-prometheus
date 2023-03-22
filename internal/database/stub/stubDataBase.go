package stub

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
)

type StubDataBase struct{}

func (s *StubDataBase) UpdateRecords(context.Context, []*database.DBRecord) error {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadRecord(context.Context, string, string) (*database.DBRecord, error) {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadAll(context.Context) ([]*database.DBRecord, error) {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) Ping(context.Context) error {
	return nil
}

func (s *StubDataBase) Close() error {
	return nil
}
