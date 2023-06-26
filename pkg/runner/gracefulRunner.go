package runner

import (
	"context"
	"fmt"
)

type GracefulRunner struct {
	starter Runner
	stopper Stopper
	done    chan struct{}
	exit    chan struct{}
	err     chan error
}

func NewGracefulRunner(runner Runner) *GracefulRunner {
	gr := &GracefulRunner{
		starter: runner,
		done:    make(chan struct{}),
		exit:    make(chan struct{}),
		err:     make(chan error, 1),
	}

	stopper, ok := runner.(Stopper)
	if ok {
		gr.stopper = stopper
	}

	return gr
}

func (gr *GracefulRunner) Start(ctx context.Context) {
	stopCtx, cancel := context.WithCancel(ctx)

	go func() {
		gr.err <- gr.starter.Start(stopCtx)
		close(gr.exit)
	}()

	go func() {
		<-gr.done
		cancel()
	}()
}

func (gr *GracefulRunner) Stop(ctx context.Context) error {
	close(gr.done)

	if gr.stopper != nil {
		if err := gr.stopper.Stop(ctx); err != nil {
			return err
		}
	}

	select {
	case <-gr.exit:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("runner can't stop: %w", ctx.Err())
	}
}

func (gr *GracefulRunner) Error() <-chan error {
	return gr.err
}
