package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

func TestPeriodicWorker_CloseContext(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())

	worker := NewPeriodicWorker(1*time.Millisecond, func(context.Context) error {
		wasCalled = true
		return nil
	})

	cancel()
	_ = worker.Start(ctx)
	assert.False(t, wasCalled)
}

func TestPeriodicWorker_SuccessCall(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := NewPeriodicWorker(1*time.Millisecond, func(context.Context) error {
		if !wasCalled {
			wasCalled = true
			return test.ErrTest
		}

		cancel()
		return nil
	})

	_ = worker.Start(ctx)
	assert.True(t, wasCalled)
}
