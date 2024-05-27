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
	Start() (chan *I, chan *O, chan error, error)
	Stop() error
	AddPool(chan *I, chan error, int, WorkerPool[I, O]) error
	GetPool() WorkerPool[I, O]
}

// PoolSupervisorNode represents a supervisor node that manages a pool of workers.
type PoolSupervisorNode[I, O any] struct {
	node
	inChan       chan *I
	outChan      chan *O
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
func (s *PoolSupervisorNode[I, O]) AddPool(inputCh chan *I, eventOutChan chan error, buffSize int, pool WorkerPool[I, O]) error {
	if s.nodeType != NodeTypeDefault && s.nodeType != NodeTypePoolSupervisor {
		return NewError[error](s.ID, ErrorLevelError, fmt.Sprintf("Cannot add a worker pool to a %s node", nodeTypeNames[s.nodeType]), nil, nil)
	}

	s.inChan = inputCh
	s.outChan = make(chan *O, buffSize)
	s.eventOutChan = eventOutChan
	s.pool = pool
	return nil
}

// GetName returns the name of the supervisor node.
func (s *PoolSupervisorNode[I, O]) GetName() string {
	return s.name
}

// GetPool returns the worker pool.
func (s *PoolSupervisorNode[I, O]) GetPool() WorkerPool[I, O] {
	return s.pool
}

// Start starts the supervisor and its worker pool.
func (s *PoolSupervisorNode[I, O]) Start() (chan *I, chan *O, chan error, error) {
	if s.inChan == nil || s.outChan == nil || s.eventInChan == nil {
		return nil, nil, nil, errors.New("input, output, or error channel is not initialized")
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	s.cancelFunc = cancelFunc

	s.pool.Start(s.inChan, s.outChan, s.eventInChan)

	go func() {
		for {
			select {
			case event := <-s.eventInChan:
				s.Handle(event)
			case <-ctx.Done():
				return
			}
		}
	}()

	return s.inChan, s.outChan, s.eventInChan, nil
}

// Handle processes an error event.
func (s *PoolSupervisorNode[I, O]) Handle(event error) {
	var customErr *Error[I]
	if errors.As(event, &customErr) {
		if customErr.Level == ErrorLevelCritical {
			err := s.pool.RemoveWorker(customErr.Node)
			if err != nil {
				s.eventOutChan <- NewError[error](s.ID, ErrorLevelCritical, "worker removal error", err, nil)
			}
			// Replace the worker
			s.pool.AddWorkers(1)
		}
		s.inChan <- customErr.InputItem
		fmt.Printf("Resubmitted input: %v\n", customErr.InputItem)
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
		close(s.outChan)
		close(s.eventInChan)
		return nil
	}
	return errors.New("supervisor not running")
}
