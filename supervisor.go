package soop

import (
	"context"
	"errors"
	"fmt"
)

// EventHandler processes an error event.
type EventHandler func(error) error

// Supervisor represents a supervisor that manages a pool of workers.
type Supervisor[I, O any] interface {
	Start() (chan *I, chan *O, chan error, error)
	Stop() error
}

// SupervisorNode represents a supervisor node that manages a pool of workers.
type SupervisorNode[I, O any] struct {
	node
	inChan     chan *I
	outChan    chan *O
	evOutChan  chan error
	handler    EventHandler
	pool       WorkerPool[I, O]
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewSupervisorNode creates a new supervisor node.
func NewSupervisorNode[I, O any](ctx context.Context, name string, handler EventHandler, pool WorkerPool[I, O]) Supervisor[I, O] {
	return &SupervisorNode[I, O]{
		node:      newNode(ctx, name, NodeTypeSupervisor, nil),
		inChan:    make(chan *I),
		outChan:   make(chan *O),
		evOutChan: make(chan error),
		handler:   handler,
		pool:      pool,
		ctx:       ctx,
	}
}

// Start starts the supervisor and its worker pool.
func (s *SupervisorNode[I, O]) Start() (chan *I, chan *O, chan error, error) {
	if s.inChan == nil || s.outChan == nil || s.evOutChan == nil {
		return nil, nil, nil, errors.New("input, output, or error channel is not initialized")
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	s.cancelFunc = cancelFunc

	s.pool.Start(s.inChan, s.outChan, s.evOutChan)

	go func() {
		for {
			select {
			case err := <-s.evOutChan:
				s.Handle(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return s.inChan, s.outChan, s.evOutChan, nil
}

// Handle processes an error event.
func (s *SupervisorNode[I, O]) Handle(event error) {
	var customErr *Error[I]
	if errors.As(event, &customErr) {
		if customErr.Level == ErrorLevelCritical {
			err := s.pool.RemoveWorker(customErr.Node)
			if err != nil {
				fmt.Printf("Failed to remove worker: %v\n", err)
			}
			// Replace the worker after removal
			s.pool.AddWorkers(1)
		}
		s.inChan <- customErr.InputItem
	}
	if err := s.handler(event); err != nil {
		s.evOutChan <- err
	}
}

// Stop stops the supervisor and its worker pool.
func (s *SupervisorNode[I, O]) Stop() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil // Set to nil to prevent multiple cancels
		close(s.inChan)
		close(s.outChan)
		close(s.evOutChan)
		return nil
	}
	return errors.New("supervisor not running")
}
