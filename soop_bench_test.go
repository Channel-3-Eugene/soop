package soop

// import (
// 	"context"
// 	"testing"
// 	"time"
// )

// const BENCH_COUNT = 10000

// type BenchmarkWorker struct{}

// func (w *BenchmarkWorker) Handle(input *TestInputType) (*TestOutputType, error) {
// 	// Simulate work
// 	return &TestOutputType{n: input.c}, nil
// }

// func benchmarkSupervisor(b *testing.B, workerCount, inputCount int) {
// 	ctx := context.Background()
// 	inChan := make(chan Envelope[TestInputType])
// 	outChan := make(chan TestOutputType)
// 	evOutChan := make(chan error)
// 	pool := NewWorkerPool(ctx, "pool", workerCount, workerCount, func() Worker[TestInputType, TestOutputType] {
// 		return &BenchmarkWorker{}
// 	})
// 	supervisor := NewSupervisorNode(ctx, "supervisor", &inChan, &outChan, &evOutChan, nil, pool)
// 	inputCh, outputCh, errorCh, err := supervisor.Start()
// 	if err != nil {
// 		b.Fatal("Failed to start supervisor:", err)
// 	}

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {

// 		var c int
// 		done := make(chan struct{})
// 		timeout := time.After(TIMEOUT)

// 		go func() {
// 			defer close(done)
// 			for {
// 				select {
// 				case _, ok := <-outputCh:
// 					if !ok {
// 						return
// 					}
// 					c++
// 					if c == inputCount {
// 						return
// 					}
// 				case err := <-errorCh:
// 					b.Errorf("Error received from worker on error channel: %v", err)
// 				case <-timeout:
// 					b.Errorf("Timed out waiting for response")
// 					return
// 				}
// 			}
// 		}()

// 		for i := 0; i < inputCount; i++ {
// 			inputCh <- Envelope[TestInputType]{Item: TestInputType{i}}
// 		}

// 		<-done
// 	}

// 	supervisor.Stop()
// }

// func BenchmarkSupervisor_3Workers_10000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 3, BENCH_COUNT)
// }

// func BenchmarkSupervisor_10Workers_10000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 10, BENCH_COUNT)
// }

// func BenchmarkSupervisor_50Workers_10000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 50, BENCH_COUNT)
// }

// func BenchmarkSupervisor_100Workers_10000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 100, BENCH_COUNT)
// }

// func BenchmarkSupervisor_3Workers_50000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 3, 50000)
// }

// func BenchmarkSupervisor_10Workers_50000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 10, 50000)
// }

// func BenchmarkSupervisor_50Workers_50000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 50, 50000)
// }

// func BenchmarkSupervisor_100Workers_50000Inputs(b *testing.B) {
// 	benchmarkSupervisor(b, 100, 50000)
// }
