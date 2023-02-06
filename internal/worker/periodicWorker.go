package worker

import (
	"context"
	"go-yandex-aka-prometheus/internal/logger"
	"time"
)

type PeriodicWorkerConfig struct {
	Duration time.Duration
}

type PeriodicWorker struct {
	duration time.Duration
	workFunc func(ctx context.Context) error
}

func NewPeriodicWorker(config PeriodicWorkerConfig, workFunc func(ctx context.Context) error) PeriodicWorker {
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
			err := w.workFunc(ctx)
			if err != nil {
				logger.ErrorFormat("Periodic worker error: %v", err.Error())
			}
		case <-ctx.Done():
			logger.ErrorFormat("Periodic worker canceled")
			break
		}
	}
}
