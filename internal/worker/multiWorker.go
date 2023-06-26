package worker

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/pkg/runner"
	"golang.org/x/sync/errgroup"
)

type multiWorker struct {
	runners []runner.Runner
}

func NewMultiWorker(runners ...runner.Runner) *multiWorker {
	return &multiWorker{
		runners: runners,
	}
}

func (m *multiWorker) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < len(m.runners); i++ {
		num := i
		eg.Go(func() error {
			return m.runners[num].Start(ctx)
		})
	}

	return eg.Wait()
}
