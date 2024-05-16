package soop

import (
	"context"
	"fmt"
	"sync"
)

// Worker interface with generic input and output types
type Worker[I, O any] interface {
	Handle(ctx context.Context, in <-chan I, out chan<- O, errCh chan<- error)
}

// Supervisor manages a pool of workers using generics
type Supervisor[I, O any] struct {
	poolSize int
	inputCh  chan I
	outputCh chan O
	errorCh  chan error
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewSupervisor creates a new supervisor
func NewSupervisor[I, O any](ctx context.Context, poolSize int) *Supervisor[I, O] {
	ctx, cancel := context.WithCancel(ctx)
	return &Supervisor[I, O]{
		poolSize: poolSize,
		inputCh:  make(chan I),
		outputCh: make(chan O),
		errorCh:  make(chan error, 100), // Buffer size can be adjusted
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start initializes the supervisor with a pool of workers and returns shared input/output channels
func (s *Supervisor[I, O]) Start(workerFactory func() Worker[I, O]) (chan I, chan O, chan error, error) {
	if s.poolSize <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid pool size: %d", s.poolSize)
	}

	for i := 0; i < s.poolSize; i++ {
		worker := workerFactory()
		s.wg.Add(1) // Add 1 to the WaitGroup for each worker
		go func() {
			defer s.wg.Done() // Ensure Done is called when the goroutine exits
			worker.Handle(s.ctx, s.inputCh, s.outputCh, s.errorCh)
		}()
	}
	return s.inputCh, s.outputCh, s.errorCh, nil
}

// Stop gracefully stops the supervisor and its workers
func (s *Supervisor[I, O]) Stop() error {
	s.cancel()
	s.wg.Wait()
	close(s.outputCh)
	close(s.errorCh)
	return nil
}
