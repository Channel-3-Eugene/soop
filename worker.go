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
	Start(context.Context, chan *I, chan *O, chan error)
}

// WorkerNode represents a worker node that processes inputs and produces outputs.
type WorkerNode[I, O any] struct {
	node
	handler InputHandler[I, O]
}

// NewWorkerNode creates a new worker node.
func NewWorkerNode[I, O any](ctx context.Context, name string, handler InputHandler[I, O]) Worker[I, O] {
	worker := &WorkerNode[I, O]{
		handler: handler,
	}
	worker.node = newNode(ctx, name, NodeTypeWorker, func() Node {
		return NewWorkerNode(ctx, name, handler).(*WorkerNode[I, O])
	})
	return worker
}

// Handle processes an input and returns an output or an error.
func (w *WorkerNode[I, O]) Handle(input *I) (O, error) {
	return w.handler(input)
}

// Start starts the worker and processes inputs from the input channel.
func (w *WorkerNode[I, O]) Start(ctx context.Context, inChan chan *I, outChan chan *O, errChan chan error) {
	go func() {
		fmt.Printf("%s started\n", w.name)
		defer func() {
			if r := recover(); r != nil {
				switch r := r.(type) {
				case *I:
					fmt.Printf("%s panicked with (I) %#v\n", w.name, r)
					errChan <- NewError(w.ID, ErrorLevelCritical, "worker panicked", fmt.Errorf("%v", r), r)
				case error:
					fmt.Printf("%s panicked with (error) %#v\n", w.name, r)
					errChan <- NewError(w.ID, ErrorLevelCritical, "worker panicked", r, (*I)(nil))
				default:
					fmt.Printf("%s panicked with (default) %#v\n", w.name, r)
					errChan <- NewError(w.ID, ErrorLevelCritical, "worker panicked", fmt.Errorf("%v", r), (*I)(nil))
				}
				return
			}
		}()

		for {
			select {
			case input, ok := <-inChan:
				if !ok {
					return
				}
				func() {
					output, err := w.Handle(input)
					if err != nil {
						errChan <- NewError(w.ID, ErrorLevelError, "processing error", err, input)
					} else {
						fmt.Printf("%s output %#v\n", w.name, output)
						outChan <- &output
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
}
