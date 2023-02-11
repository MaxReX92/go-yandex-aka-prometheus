package worker

import (
	"context"
	"go-yandex-aka-prometheus/internal/logger"
	"time"
)

type PeriodicWorker struct {
	workFunc func(ctx context.Context) error
}

func NewPeriodicWorker(workFunc func(ctx context.Context) error) PeriodicWorker {
	return PeriodicWorker{
		workFunc: workFunc,
	}
}

func (w *PeriodicWorker) StartWork(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := w.workFunc(ctx)
			if err != nil {
				logger.ErrorFormat("Periodic worker error: %v", err.Error())
			}
		case <-ctx.Done():
			logger.ErrorFormat("Periodic worker canceled")
			return
		}
	}
}
