package worker_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"{{ cookiecutter.import_path }}/internal/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolBasic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := worker.NewPool(ctx, 2, slog.Default())
	pool.Start()

	var mu sync.Mutex
	results := make([]int, 0, 5)

	for i := 0; i < 5; i++ {
		ok := pool.Submit(func(ctx context.Context) error {
			mu.Lock()
			results = append(results, i)
			mu.Unlock()
			return nil
		})
		require.True(t, ok)
	}

	pool.Stop()

	require.Len(t, results, 5)
}

func TestPoolError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := worker.NewPool(ctx, 1, slog.Default())
	pool.Start()

	errs := make(chan error, 10)
	go func() {
		for err := range pool.Results() {
			errs <- err
		}
	}()

	pool.Submit(func(ctx context.Context) error {
		return errors.New("test error")
	})

	pool.Stop()

	select {
	case err := <-errs:
		assert.Contains(t, err.Error(), "test error")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error result")
	}
}

func TestPoolCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	pool := worker.NewPool(ctx, 1, slog.Default())
	pool.Start()

	// Submit a long-running task
	pool.Submit(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	// Cancel the parent context
	cancel()

	// Wait a bit for the task to notice cancellation
	time.Sleep(100 * time.Millisecond)

	pool.Stop()
}

func TestPoolDone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := worker.NewPool(ctx, 1, slog.Default())
	pool.Start()

	// Done channel should not be closed yet
	select {
	case <-pool.Done():
		t.Fatal("Done channel should not be closed before Stop")
	default:
	}

	pool.Stop()

	// Done channel should be closed now
	select {
	case <-pool.Done():
		// Expected
	default:
		t.Fatal("Done channel should be closed after Stop")
	}
}
