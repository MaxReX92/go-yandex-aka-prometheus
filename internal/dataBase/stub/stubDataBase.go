package stub

import "context"

type StubDataBase struct {
}

func (s *StubDataBase) Ping(ctx context.Context) error {
	return nil
}

func (s *StubDataBase) Close() error {
	return nil
}
