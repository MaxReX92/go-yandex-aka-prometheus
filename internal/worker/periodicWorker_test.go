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
	err := worker.Start(ctx)
	assert.NoError(t, err)
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

	err := worker.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, wasCalled)
}
