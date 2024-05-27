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
	inChan       chan *I
	outChan      chan *O
	eventInChan  chan error
	eventOutChan chan error
	handler      EventHandler
	pool         WorkerPool[I, O]
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

// NewSupervisorNode creates a new supervisor node.
func NewSupervisorNode[I, O any](ctx context.Context, name string, inputCh chan *I, buffSize int, handler EventHandler, pool WorkerPool[I, O]) Supervisor[I, O] {
	return &SupervisorNode[I, O]{
		node:         newNode(ctx, name, NodeTypeSupervisor, nil),
		inChan:       inputCh,
		outChan:      make(chan *O, buffSize),
		eventInChan:  make(chan error, 10),
		eventOutChan: nil,
		handler:      handler,
		pool:         pool,
		ctx:          ctx,
	}
}

// Start starts the supervisor and its worker pool.
func (s *SupervisorNode[I, O]) Start() (chan *I, chan *O, chan error, error) {
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
				fmt.Printf("\nSupervisor processing event: %v\n\n", event)
				s.Handle(event)
				fmt.Printf("\nSupervisor processed event: %v\n\n", event)
			case <-ctx.Done():
				return
			}
		}
	}()

	return s.inChan, s.outChan, s.eventInChan, nil
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
func (s *SupervisorNode[I, O]) Stop() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil // Set to nil to prevent multiple cancels
		// Not closing s.inChan and s.eventInChan because they are owned by other processes
		close(s.outChan)
		close(s.eventInChan)
		return nil
	}
	return errors.New("supervisor not running")
}
