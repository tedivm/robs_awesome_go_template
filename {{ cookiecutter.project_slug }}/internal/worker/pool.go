package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Task represents a unit of work.
type Task func(ctx context.Context) error

// Pool is a worker pool for CPU-bound tasks.
type Pool struct {
	ctx     context.Context
	cancel  context.CancelFunc
	size    int
	logger  *slog.Logger
	tasks   chan Task
	results chan error
	wg      sync.WaitGroup
	done    chan struct{}
}

// NewPool creates a new worker pool with the given size.
func NewPool(ctx context.Context, size int, logger *slog.Logger) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	return &Pool{
		ctx:     ctx,
		cancel:  cancel,
		size:    size,
		logger:  logger,
		tasks:   make(chan Task, size*10),
		results: make(chan error, size*10),
		done:    make(chan struct{}),
	}
}

// Start launches worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.work()
	}
}

func (p *Pool) work() {
	defer p.wg.Done()
	for task := range p.tasks {
		if err := task(p.ctx); err != nil {
			err = fmt.Errorf("task failed: %w", err)
			p.logger.Error("worker task failed", "error", err)
			select {
			case p.results <- err:
			case <-p.ctx.Done():
				return
			default:
				// Results channel is full and no one is reading.
				// Error was logged above; drop to prevent blocking shutdown.
			}
		}
	}
}

// Submit adds a task to the pool. Returns false if the pool is shutting down.
func (p *Pool) Submit(task Task) bool {
	select {
	case p.tasks <- task:
		return true
	case <-p.ctx.Done():
		return false
	}
}

// Results returns a read-only channel for task errors.
func (p *Pool) Results() <-chan error {
	return p.results
}

// Stop gracefully shuts down the pool, waiting for in-flight tasks to complete.
func (p *Pool) Stop() {
	// Close tasks channel first to stop workers from accepting new work
	close(p.tasks)
	// Wait for in-flight tasks to finish
	p.wg.Wait()
	// Cancel context to signal any blocked operations
	p.cancel()
	// Close output channels
	close(p.results)
	close(p.done)
}

// Done returns a channel that closes when the pool has fully stopped.
func (p *Pool) Done() <-chan struct{} {
	return p.done
}
