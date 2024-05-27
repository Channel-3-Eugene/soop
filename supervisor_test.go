package soop

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func eventHandler(e error) error {
	return e
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

	inChan := make(chan *TestInputType, 100)
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
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

	inChan := make(chan *TestInputType, 100)
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
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

	close(inChan)
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

	inChan := make(chan *TestInputType, 100)
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
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
	close(inChan)
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

	inChan := make(chan *TestInputType, 100)
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
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
	close(inChan)
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

	inChan := make(chan *TestInputType, 100)
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
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
	close(inChan)
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}

func TestSupervisor_RestartsWorkers(t *testing.T) {
	const COUNT = 12
	const WORKERS = 2

	ctx := context.Background()

	// Create a worker pool with a handler that panics randomly
	pool, err := NewWorkerPool(ctx, WORKERS-1, WORKERS+1, func(id uint64) *WorkerNode[TestInputType, TestOutputType] {
		return NewWorkerNode(ctx, "node", func(input *TestInputType) (TestOutputType, error) {
			n, _ := rand.Int(rand.Reader, big.NewInt(5))
			if n.Int64() == 0 {
				panic(input)
			}
			return TestOutputType{n: input.c}, nil
		}).(*WorkerNode[TestInputType, TestOutputType])
	})
	assert.NoError(t, err, "Worker pool should be initialized without error")
	assert.NotNil(t, pool, "Worker pool should not be nil")

	inChan := make(chan *TestInputType, 10)
	assert.NotNil(t, inChan, "New input channel should not be nil")
	supervisor := NewSupervisorNode(ctx, "supervisor", inChan, 100, nil, pool)
	assert.NotNil(t, supervisor, "Supervisor should be initialized")

	inputCh, outputCh, _, err := supervisor.Start()
	assert.NoError(t, err, "Supervisor should start without error")
	assert.NotNil(t, inputCh, "Input channel should not be nil")
	assert.NotNil(t, outputCh, "Output channel should not be nil")

	// Send inputs in a separate goroutine
	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- &TestInputType{c: i}
			fmt.Printf("Sent input %#v\n", TestInputType{c: i})
		}
	}()

	recdItems := 0
	for output := range outputCh {
		recdItems++
		fmt.Printf("Received output %#v\n", output)
		if recdItems == COUNT {
			break
		}
	}

	assert.Equal(t, COUNT, recdItems, "Expected %d items to be received, got %d", recdItems)

	// Ensure graceful shutdown
	close(inChan)
	err = supervisor.Stop()
	assert.NoError(t, err, "Supervisor should stop without error")
}
