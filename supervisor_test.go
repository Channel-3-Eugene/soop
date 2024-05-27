package soop

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Define sHandler
func sHandler(err error) error {
	return err
}

// Test for zero workers
func TestSupervisor_ZeroWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := NewWorkerPool(ctx, 0, 0, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "worker-test", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.Error(t, err, "Expected error for zero workers")
}

func TestSupervisor_NegativeWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := NewWorkerPool(ctx, -1, -1, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "worker-test", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.Error(t, err, "Expected error for negative workers")
	assert.EqualError(t, err, "min workers must be greater than zero", "Expected 'min workers must be greater than zero' error")
}

func TestSupervisor_Initialization(t *testing.T) {
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

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")
}

func TestSupervisor_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	COUNT := 10

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	// Send some inputs to process
	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	// drain output channel
	recdItems := 0
	for range outputCh {
		recdItems++
		if recdItems == COUNT {
			break
		}
	}

	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_Process(t *testing.T) {
	const COUNT = 10000
	const WORKERS = 5

	// Time the actual test run
	start := time.Now()
	defer func() {
		fmt.Printf("Test took %v for %d iterations\n", time.Since(start), COUNT)
		fmt.Printf("Using %d workers\n", WORKERS)
		fmt.Printf("Average processing time per item: %v\n", time.Since(start)/time.Duration(COUNT))
		fmt.Printf("Iterations per second: %v\n", float64(COUNT)/time.Since(start).Seconds())
	}()

	ctx := context.Background()

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	recdItems := 0
	for range outputCh {
		recdItems++
		if recdItems == COUNT {
			break
		}
	}

	assert.Equal(t, COUNT, recdItems, "Expected %d items to be received, got %d", COUNT, recdItems)

	// Ensure graceful shutdown
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_HighConcurrency(t *testing.T) {
	var (
		COUNT   = int(math.Pow(2, 16))
		WORKERS = int(math.Pow(2, 6))
	)

	// Time the actual test run
	start := time.Now()
	defer func() {
		fmt.Printf("Test took %v for %d iterations\n", time.Since(start), COUNT)
		fmt.Printf("Using %d workers\n", WORKERS)
		fmt.Printf("Average processing time per item: %v\n", time.Since(start)/time.Duration(COUNT))
		fmt.Printf("Iterations per second: %v\n", float64(COUNT)/time.Since(start).Seconds())
	}()

	ctx := context.Background()

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	// Send inputs
	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	// Receive outputs
	recdItems := 0
	for range outputCh {
		recdItems++
		if recdItems == COUNT {
			break
		}
	}

	assert.Equal(t, COUNT, recdItems, "Expected %d items to be received, got %d", COUNT, recdItems)

	// Ensure graceful shutdown
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_MixedWorkloads(t *testing.T) {
	const COUNT = 100
	const WORKERS = 5

	ctx := context.Background()

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			if input.c%2 == 0 {
				return TestOutputType{n: input.c}, nil
			} else {
				return TestOutputType{}, fmt.Errorf("simulated error")
			}
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, errCh, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	recdItems := 0
	errorCount := 0
	for recdItems+errorCount < COUNT {
		select {
		case output := <-outputCh:
			assert.NotNil(t, output, "Output should not be nil")
			recdItems++
		case err := <-errCh:
			assert.Error(t, err, "Expected an error")
			errorCount++
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for items")
		}
	}

	assert.True(t, recdItems > 0, "Expected some items to be received")
	assert.True(t, errorCount > 0, "Expected some errors")
	assert.Equal(t, COUNT, recdItems+errorCount, "Expected %d items to be processed in total, got %d", COUNT, recdItems+errorCount)

	// Ensure graceful shutdown
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_WithDelays(t *testing.T) {
	const COUNT = 100
	const WORKERS = 5

	ctx := context.Background()

	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		w := NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			time.Sleep(time.Duration(input.c%10) * time.Millisecond) // Simulate work delay
			return TestOutputType{n: input.c}, nil
		})
		return w.(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", sHandler, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	recdItems := 0
	for range outputCh {
		recdItems++
		if recdItems == COUNT {
			break
		}
	}

	assert.Equal(t, COUNT, recdItems, "Expected %d items to be received, got %d", COUNT, recdItems)

	// Ensure graceful shutdown
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_RestartsWorkers(t *testing.T) {
	const COUNT = 12
	const WORKERS = 2

	ctx := context.Background()

	// Create a worker pool with a handler that panics for every 10th input
	pool, err := NewWorkerPool(ctx, WORKERS, WORKERS, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		return NewWorkerNode(ctx, fmt.Sprintf("node-%d", id), func(input *TestInputType) (TestOutputType, error) {
			if input.c%3 == 0 {
				panic("simulated panic")
			}
			return TestOutputType{n: input.c}, nil
		}).(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	supervisor := NewSupervisorNode(ctx, "supervisor", func(err error) error {
		// Log the error without failing the test
		t.Logf("Supervisor handled error: %v", err)
		return nil
	}, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")

	// Send inputs in a separate goroutine
	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
		}
	}()

	recdItems := 0
	for range outputCh {
		recdItems++
		println(recdItems, " items received")
		if recdItems == COUNT {
			break
		}
	}

	assert.Equal(t, COUNT, recdItems, "Expected %d items to be received, got %d", recdItems)

	// Ensure graceful shutdown
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}
