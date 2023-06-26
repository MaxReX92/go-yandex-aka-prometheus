package worker

import (
	"context"
	"time"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
)

// PeriodicWorker will start function at period of time.
type PeriodicWorker struct {
	interval time.Duration
	workFunc func(ctx context.Context) error
}

// NewPeriodicWorker create new instance of PeriodicWorker.
func NewPeriodicWorker(interval time.Duration, workFunc func(ctx context.Context) error) PeriodicWorker {
	return PeriodicWorker{
		interval: interval,
		workFunc: workFunc,
	}
}

// StartWork start worker function at period of time.
func (w *PeriodicWorker) Start(ctx context.Context) error {
	// parent context is not used consciously, or graceful shutdown
	actionContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := w.workFunc(actionContext)
			if err != nil {
				logger.ErrorFormat("periodic worker error: %v", err)
			}
		case <-ctx.Done():
			logger.ErrorFormat("periodic worker canceled")
			return ctx.Err()
		}
	}
}
