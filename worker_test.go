package soop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Existing tests

func TestWorkerNode_ProcessInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inChan := make(chan *int)
	outChan := make(chan *int)
	errChan := make(chan error)

	handler := func(input *int) (int, error) {
		return *input * 2, nil
	}

	worker := NewWorkerNode(ctx, "testWorker", handler)
	worker.Start(ctx, inChan, outChan, errChan)

	// Send input
	input := 5
	inChan <- &input

	// Receive output
	select {
	case output := <-outChan:
		assert.Equal(t, 10, *output, "expected 10, got %d", *output)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for output")
	}
}

func TestWorkerNode_HandleError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inChan := make(chan *int)
	outChan := make(chan *int)
	errChan := make(chan error)

	handler := func(input *int) (int, error) {
		return 0, errors.New("handler error")
	}

	worker := NewWorkerNode(ctx, "testWorker", handler)
	worker.Start(ctx, inChan, outChan, errChan)

	// Send input
	input := 5
	inChan <- &input

	// Receive error
	select {
	case err := <-errChan:
		assert.NotNil(t, err, "expected error, got nil")
		assert.ErrorContains(t, err, "handler error", "expected handler error, got %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestWorkerNode_HandlePanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inChan := make(chan *int)
	outChan := make(chan *int)
	errChan := make(chan error)

	handler := func(input *int) (int, error) {
		panic("something went wrong")
	}

	worker := NewWorkerNode(ctx, "testWorker", handler)
	worker.Start(ctx, inChan, outChan, errChan)

	// Send input
	input := 5
	inChan <- &input

	// Receive panic as error
	select {
	case err := <-errChan:
		assert.NotNil(t, err, "expected panic error, got nil")
		assert.ErrorContains(t, err, "panicked", "expected panic error, got %v", err.Error())
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for panic error")
	}
}

func TestWorkerNode_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inChan := make(chan *int)
	outChan := make(chan *int)
	errChan := make(chan error)

	handler := func(input *int) (int, error) {
		return *input * 2, nil
	}

	worker := NewWorkerNode(ctx, "testWorker", handler)
	worker.Start(ctx, inChan, outChan, errChan)

	// Cancel the context
	cancel()

	// Close the input channel
	close(inChan)

	// Ensure the worker exits
	select {
	case _, ok := <-inChan:
		assert.False(t, ok, "expected inChan to be closed")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for worker to exit")
	}
}

func TestWorkerPool_Initialization(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")
}

func TestWorkerPool_Size(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	assert.Equal(t, 5, pool.Size(), "Expected pool size to be 5")
}

func TestWorkerPool_MinWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	assert.Equal(t, 3, pool.MinWorkers(), "Expected minimum workers to be 3")
}

func TestWorkerPool_MaxWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	assert.Equal(t, 5, pool.MaxWorkers(), "Expected maximum workers to be 5")
}

func TestWorkerPool_AddWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	pool.AddWorkers(2)
	assert.Equal(t, 7, pool.Size(), "Expected pool size to be 7 after adding 2 workers")
}

func TestWorkerPool_RemoveWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	err = pool.RemoveWorker(1)
	assert.NoError(t, err, "Expected no error when removing existing worker")
	assert.Equal(t, 4, pool.Size(), "Expected pool size to be 4 after removing 1 worker")

	err = pool.RemoveWorker(10)
	assert.Error(t, err, "Expected error when removing non-existing worker")
}

func TestWorkerPool_RemoveWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewWorkerPool(ctx, 3, 5, func(id uint64) *WorkerNode[int, int] {
		return NewWorkerNode(ctx, "node", func(input *int) (int, error) {
			return *input, nil
		}).(*WorkerNode[int, int])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	err = pool.RemoveWorkers(2)
	assert.NoError(t, err, "Expected no error when removing 2 workers")
	assert.Equal(t, 3, pool.Size(), "Expected pool size to be 3 after removing 2 workers")

	err = pool.RemoveWorkers(5)
	assert.Error(t, err, "Expected error when removing more workers than exist in the pool")
}
