package stub

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/database"
)

type StubDataBase struct{}

func (s *StubDataBase) UpdateRecords(ctx context.Context, records []*database.DBRecord) error {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadRecord(ctx context.Context, metricType string, metricName string) (*database.DBRecord, error) {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) ReadAll(ctx context.Context) ([]*database.DBRecord, error) {
	// TODO implement me
	panic("implement me")
}

func (s *StubDataBase) Ping(ctx context.Context) error {
	return nil
}

func (s *StubDataBase) Close() error {
	return nil
}
