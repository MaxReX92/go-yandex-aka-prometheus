package runner

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type multiWorker struct {
	runners []Runner
}

func NewMultiWorker(runners ...Runner) *multiWorker {
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

func (m *multiWorker) Stop(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < len(m.runners); i++ {
		num := i
		eg.Go(func() error {
			stopper, ok := m.runners[num].(Stopper)
			if ok {
				return stopper.Stop(ctx)
			}

			return nil
		})
	}

	return eg.Wait()
}
