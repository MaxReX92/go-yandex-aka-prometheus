package worker

import (
	"context"
	"time"
)

type PeriodicWorkerConfig struct {
	Duration time.Duration
}

type PeriodicWorker struct {
	duration time.Duration
	workFunc func() error
}

func NewPeriodicWorker(config PeriodicWorkerConfig, workFunc func() error) PeriodicWorker {
	return PeriodicWorker{
		duration: config.Duration,
		workFunc: workFunc,
	}
}

func (w *PeriodicWorker) StartWork(ctx context.Context) {
	ticker := time.NewTicker(w.duration)

	for true {
		select {
		case <-ticker.C:
			err := w.workFunc()
			if err != nil {
				// TODO: log
			}
		case <-ctx.Done():
			// TODO: log
			break
		}
	}
}