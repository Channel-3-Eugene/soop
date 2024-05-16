package soop

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const COUNT = 10000
const TIMEOUT = 1 * time.Second
const WORKERS = 3

type TestInputType struct{ c int }
type TestOutputType struct{ n int }

type ProcessWorker struct{}

func (w *ProcessWorker) Handle(ctx context.Context, in <-chan TestInputType, out chan<- TestOutputType, errCh chan<- error) {
	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return
			}
			// Simulate work
			select {
			case out <- TestOutputType{msg.c}:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		}
	}
}

func TestSupervisor_ZeroWorkers(t *testing.T) {
	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, 0)
	_, _, _, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ProcessWorker{}
	})
	assert.Error(t, err, "Expected error for zero workers")
}

func TestSupervisor_NegativeWorkers(t *testing.T) {
	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, -1)
	_, _, _, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ProcessWorker{}
	})
	assert.Error(t, err, "Expected error for negative workers")
}

func TestSupervisor_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, WORKERS)
	inputCh, outputCh, errorCh, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ProcessWorker{}
	})
	if err != nil {
		t.Fatal("Failed to start supervisor:", err)
	}

	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- TestInputType{i}
		}
	}()

	var c int
	timeout := time.After(5 * time.Millisecond)

	go func() {
		for {
			select {
			case _, ok := <-outputCh:
				if !ok {
					return
				}
				c++
			case err := <-errorCh:
				if err != context.Canceled {
					assert.NoError(t, err, "Error received from worker on error channel")
				}
			case <-timeout:
				return
			}
		}
	}()

	// Cancel the context to initiate shutdown
	cancel()

	err = supervisor.Stop()
	if err != nil {
		t.Fatal("Failed to stop supervisor:", err)
	}
}

type ErrorWorker struct{}

func (w *ErrorWorker) Handle(ctx context.Context, in <-chan TestInputType, out chan<- TestOutputType, errCh chan<- error) {
	for {
		select {
		case _, ok := <-in:
			if !ok {
				return
			}
			errCh <- fmt.Errorf("simulated error")
			return
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		}
	}
}

func TestSupervisor_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, WORKERS)
	inputCh, outputCh, errorCh, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ErrorWorker{}
	})
	if err != nil {
		t.Fatal("Failed to start supervisor:", err)
	}

	go func() {
		for i := 0; i < COUNT; i++ {
			inputCh <- TestInputType{i}
		}
	}()

	var c int
	timeout := time.After(TIMEOUT)

	go func() {
		for {
			select {
			case _, ok := <-outputCh:
				if !ok {
					return
				}
				c++
				if c == COUNT {
					return
				}
			case err := <-errorCh:
				assert.Error(t, err, "Expected error from worker")
				return
			case <-timeout:
				t.Error("Timed out waiting for response")
				return
			}
		}
	}()

	err = supervisor.Stop()
	if err != nil {
		t.Fatal("Failed to stop supervisor:", err)
	}
}

func TestSupervisor_Process(t *testing.T) {
	// Time the actual test run
	start := time.Now()
	defer func() {
		fmt.Printf("Test took %v for %d iterations\n", time.Since(start), COUNT)
		fmt.Printf("Using %d workers\n", WORKERS)
		fmt.Printf("Average processing time per item: %v\n", time.Since(start)/time.Duration(COUNT))
		fmt.Printf("Iterations per second: %v\n", float64(COUNT)/time.Since(start).Seconds())
	}()

	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, WORKERS)
	inputCh, outputCh, errorCh, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ProcessWorker{}
	})
	if err != nil {
		t.Fatal("Failed to start supervisor:", err)
	}

	timeout := time.After(5 * time.Second)

	go func() {
		var c int
		var done bool

		for {
			select {
			case _, ok := <-outputCh:
				if !ok {
					t.Error("Output channel closed unexpectedly")
				}
				c++
				done = c == COUNT
			case err := <-errorCh:
				assert.NoError(t, err, "Error received from worker on error channel")
			case <-timeout:
				t.Error("Timed out waiting for response")
			}

			if done {
				break
			}
		}

		supervisor.Stop()
	}()

	for i := 0; i < COUNT; i++ {
		inputCh <- TestInputType{i}
	}
}

func TestSupervisor_HighConcurrency(t *testing.T) {
	count := COUNT * 50
	workers := WORKERS * 50

	// Time the actual test run
	start := time.Now()
	defer func() {
		fmt.Printf("Test took %v for %d iterations\n", time.Since(start), count)
		fmt.Printf("Using %d workers\n", workers)
		fmt.Printf("Average processing time per item: %v\n", time.Since(start)/time.Duration(count))
		fmt.Printf("Iterations per second: %v\n", float64(count)/time.Since(start).Seconds())
	}()

	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, workers)
	inputCh, outputCh, errorCh, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &ProcessWorker{}
	})
	if err != nil {
		t.Fatal("Failed to start supervisor:", err)
	}

	timeout := time.After(30 * time.Second)

	go func() {
		var c int
		var done bool

		for {
			select {
			case _, ok := <-outputCh:
				if !ok {
					t.Error("Output channel closed unexpectedly")
				}
				c++
				done = c == count
			case err := <-errorCh:
				assert.NoError(t, err, "Error received from worker on error channel")
			case <-timeout:
				t.Error("Timed out waiting for response")
			}

			if done {
				break
			}
		}

		supervisor.Stop()
	}()

	for i := 0; i < count; i++ {
		inputCh <- TestInputType{i}
	}
}
