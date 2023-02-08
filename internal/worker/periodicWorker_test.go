package worker

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPeriodicWorker_CloseContext(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())

	worker := NewPeriodicWorker(PeriodicWorkerConfig{Duration: 1 * time.Millisecond}, func(context.Context) error {
		wasCalled = true
		return nil
	})

	cancel()
	worker.StartWork(ctx)
	assert.False(t, wasCalled)
}

func TestPeriodicWorker_SuccessCall(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := NewPeriodicWorker(PeriodicWorkerConfig{Duration: 1 * time.Millisecond}, func(context.Context) error {
		wasCalled = true
		cancel()
		return nil
	})

	worker.StartWork(ctx)
	assert.True(t, wasCalled)
}
