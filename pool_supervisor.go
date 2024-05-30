package soop

import (
	"context"
	"errors"
	"fmt"
)

// EventHandler processes an error event.
type EventHandler func(error) error

// Supervisor represents a supervisor that manages a pool of workers.
type PoolSupervisor[I, O any] interface {
	GetID() uint64
	GetName() string
	Start() error
	Stop() error
	AddPool(WorkerPool[I, O]) error
	GetPool() WorkerPool[I, O]
	GetInChan() chan *I
	GetOutChan() chan *O
	GetEventInChan() chan error
	SetEventOutChan(chan error)
}

// PoolSupervisorNode represents a supervisor node that manages a pool of workers.
type PoolSupervisorNode[I, O any] struct {
	node
	eventInChan  chan error
	eventOutChan chan error
	handler      EventHandler
	pool         WorkerPool[I, O]
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

// NewPoolSupervisor creates a new supervisor node.
func NewPoolSupervisor[I, O any](ctx context.Context, name string, buffSize int, handler EventHandler) PoolSupervisor[I, O] {
	return &PoolSupervisorNode[I, O]{
		node:        newNode(ctx, name, NodeTypePoolSupervisor, nil),
		eventInChan: make(chan error, buffSize),
		handler:     handler,
		ctx:         ctx,
	}
}

// AddWorkerPool adds a worker pool to the supervisor node.
func (s *PoolSupervisorNode[I, O]) AddPool(pool WorkerPool[I, O]) error {
	if s.pool != nil {
		return errors.New("worker pool already initialized")
	}

	if err := pool.SetErrChan(s.eventInChan); err != nil {
		return err
	}

	s.pool = pool
	return nil
}

// GetID returns the ID of the supervisor node.
func (s *PoolSupervisorNode[I, O]) GetID() uint64 {
	return s.ID
}

// GetName returns the name of the supervisor node.
func (s *PoolSupervisorNode[I, O]) GetName() string {
	return s.name
}

// GetPool returns the worker pool.
func (s *PoolSupervisorNode[I, O]) GetPool() WorkerPool[I, O] {
	return s.pool
}

// GetInChan returns the input channel of the worker pool.
func (s *PoolSupervisorNode[I, O]) GetInChan() chan *I {
	return s.pool.GetInChan()
}

// GetOutChan returns the output channel of the worker pool.
func (s *PoolSupervisorNode[I, O]) GetOutChan() chan *O {
	return s.pool.GetOutChan()
}

// GetEventInChan returns the input event channel.
func (s *PoolSupervisorNode[I, O]) GetEventInChan() chan error {
	return s.eventInChan
}

// SetEventOutChan sets the output event channel.
func (s *PoolSupervisorNode[I, O]) SetEventOutChan(eventOutChan chan error) {
	s.eventOutChan = eventOutChan
}

// Start starts the supervisor and its worker pool.
func (s *PoolSupervisorNode[I, O]) Start() error {
	ctx, cancelFunc := context.WithCancel(s.ctx)
	s.cancelFunc = cancelFunc

	s.pool.Start()

	go func() {
		for {
			select {
			case event := <-s.eventInChan:
				fmt.Printf("Received event: %v\n", event)
				s.Handle(event)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Handle processes an error event.
func (s *PoolSupervisorNode[I, O]) Handle(event error) {
	var customErr *Error[I]
	if errors.As(event, &customErr) {
		if customErr.Level == ErrorLevelCritical {
			err := s.pool.RemoveWorker(customErr.Node)
			if err != nil {
				if s.eventOutChan != nil {
					s.eventOutChan <- NewError[error](s.ID, ErrorLevelCritical, "worker removal error", err, nil)
				} else {
					fmt.Printf("Worker %d removal error: %v\n", s.ID, err)
				}
			}
			// Replace the worker
			s.pool.AddWorkers(1)
		}
		if ch := s.pool.GetInChan(); ch != nil {
			ch <- customErr.InputItem
		} else {
			fmt.Printf("Input channel is nil!\n")
		}
		// s.pool.GetInChan() <- customErr.InputItem
		// fmt.Printf("Resubmitted input: %v\n", customErr.InputItem)
	}

	if s.eventOutChan != nil {
		s.eventOutChan <- s.handler(event)
	}
}

// Stop stops the supervisor and its worker pool.
func (s *PoolSupervisorNode[I, O]) Stop() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil // Set to nil to prevent multiple cancels
		close(s.pool.GetOutChan())
		return nil
	}
	return errors.New("supervisor not running")
}
