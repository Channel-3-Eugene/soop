package soop

import (
	"context"
	"fmt"
)

// InputHandler represents a function that processes an input and returns an output or an error.
type InputHandler[I, O any] func(*I) (O, error)

// Worker represents a worker that processes inputs and produces outputs.
type Worker[I, O any] interface {
	Handle(*I) (O, error)
	Start() error
	GetErrChan() chan error
	SetErrChan(chan error) error
}

// WorkerNode represents a worker node that processes inputs and produces outputs.
type WorkerNode[I, O any] struct {
	node
	inChan  chan *I
	outChan chan *O
	errChan chan error
	handler InputHandler[I, O]
	ctx     context.Context
}

// NewWorkerNode creates a new worker node.
func NewWorkerNode[I, O any](ctx context.Context, name string, inChan chan *I, outChan chan *O, errChan chan error, handler InputHandler[I, O]) Worker[I, O] {
	worker := &WorkerNode[I, O]{
		inChan:  inChan,
		outChan: outChan,
		errChan: errChan,
		handler: handler,
		ctx:     ctx,
	}
	worker.node = newNode(ctx, name, NodeTypeWorker, func() Node {
		return NewWorkerNode(ctx, name, inChan, outChan, errChan, handler).(*WorkerNode[I, O])
	})
	return worker
}

// Handle processes an input and returns an output or an error.
func (w *WorkerNode[I, O]) Handle(input *I) (O, error) {
	return w.handler(input)
}

// Start starts the worker and processes inputs from the input channel.
func (w *WorkerNode[I, O]) Start() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				switch r := r.(type) {
				case *I:
					err := NewError(w.ID, ErrorLevelCritical, fmt.Sprintf("worker %s panicked", w.name), fmt.Errorf("%v", r), r)
					// fmt.Printf("Worker %s panicked: %v\n", w.GetName(), err)
					if w.errChan != nil {
						w.errChan <- err
					} else {
						fmt.Printf("Worker %s error channel is nil\n", w.GetName())
					}
				case error:
					err := NewError(w.ID, ErrorLevelCritical, "worker panicked", r, (*I)(nil))
					// fmt.Printf("Worker %s panicked: %v\n", w.GetName(), err)
					if w.errChan != nil {
						w.errChan <- err
					} else {
						fmt.Printf("Worker %s error channel is nil\n", w.GetName())
					}
				default:
					err := NewError(w.ID, ErrorLevelCritical, "worker panicked", fmt.Errorf("%v", r), (*I)(nil))
					// fmt.Printf("Worker %s panicked: %v\n", w.GetName(), err)
					if w.errChan != nil {
						w.errChan <- err
					} else {
						fmt.Printf("Worker %s error channel is nil\n", w.GetName())
					}
				}
				return
			}
		}()

		for {
			select {
			case input, ok := <-w.inChan:
				if !ok {
					return
				}
				func() {
					// fmt.Printf("Worker %s received input: %v\n", w.GetName(), *input)
					output, err := w.Handle(input)
					if err != nil {
						w.errChan <- NewError(w.ID, ErrorLevelError, "processing error", err, input)
					} else {
						// fmt.Printf("Worker %s produced output: %v\n", w.GetName(), output)
						w.outChan <- &output
					}
				}()
			case <-w.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// GetErrChan returns the error channel of the worker.
func (w *WorkerNode[I, O]) GetErrChan() chan error {
	return w.errChan
}

// SetErrChan sets the error channel for the worker.
func (w *WorkerNode[I, O]) SetErrChan(errChan chan error) error {
	if errChan == nil {
		return fmt.Errorf("error channel cannot be nil")
	}
	w.errChan = errChan
	return nil
}
