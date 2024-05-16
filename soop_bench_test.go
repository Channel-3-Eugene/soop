package soop

import (
	"context"
	"testing"
	"time"
)

const BENCH_COUNT = 10000

type BenchmarkWorker struct{}

func (w *BenchmarkWorker) Handle(ctx context.Context, in <-chan TestInputType, out chan<- TestOutputType, errCh chan<- error) {
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

func benchmarkSupervisor(b *testing.B, workerCount, inputCount int) {
	ctx := context.Background()
	supervisor := NewSupervisor[TestInputType, TestOutputType](ctx, workerCount)
	inputCh, outputCh, errorCh, err := supervisor.Start(func() Worker[TestInputType, TestOutputType] {
		return &BenchmarkWorker{}
	})
	if err != nil {
		b.Fatal("Failed to start supervisor:", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {

		var c int
		done := make(chan struct{})
		timeout := time.After(TIMEOUT)

		go func() {
			defer close(done)
			for {
				select {
				case _, ok := <-outputCh:
					if !ok {
						return
					}
					c++
					if c == inputCount {
						return
					}
				case err := <-errorCh:
					b.Errorf("Error received from worker on error channel: %v", err)
				case <-timeout:
					b.Errorf("Timed out waiting for response")
					return
				}
			}
		}()

		for i := 0; i < inputCount; i++ {
			inputCh <- TestInputType{i}
		}

		<-done
	}

	supervisor.Stop()
}

func BenchmarkSupervisor_3Workers_10000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 3, BENCH_COUNT)
}

func BenchmarkSupervisor_10Workers_10000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 10, BENCH_COUNT)
}

func BenchmarkSupervisor_50Workers_10000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 50, BENCH_COUNT)
}

func BenchmarkSupervisor_100Workers_10000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 100, BENCH_COUNT)
}

func BenchmarkSupervisor_3Workers_50000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 3, 50000)
}

func BenchmarkSupervisor_10Workers_50000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 10, 50000)
}

func BenchmarkSupervisor_50Workers_50000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 50, 50000)
}

func BenchmarkSupervisor_100Workers_50000Inputs(b *testing.B) {
	benchmarkSupervisor(b, 100, 50000)
}
