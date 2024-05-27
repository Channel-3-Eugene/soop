package soop

import (
	"context"
	"fmt"
	"sync/atomic"
)

// WorkerPool interface defines the operations for a pool of workers.
type WorkerPool[I, O any] interface {
	Start(inChan chan *I, outChan chan *O, errChan chan error)
	Size() int
	MinWorkers() int
	MaxWorkers() int
	AddWorkers(n int)
	RemoveWorker(id uint64) error
	RemoveWorkers(n int) error
}

// workersMap is a map of worker IDs to worker nodes.
type workersMap[I, O any] map[uint64]*WorkerNode[I, O]

// workersFactory is a function that creates a new worker node.
type workersFactory[I, O any] func(id uint64) *WorkerNode[I, O]

// workerPool is the concrete implementation of WorkerPool.
type workerPool[I, O any] struct {
	workers    workersMap[I, O]
	factory    workersFactory[I, O]
	minWorkers int
	maxWorkers int
	inChan     chan *I
	outChan    chan *O
	errChan    chan error
	ctx        context.Context
	nextID     uint64 // Monotonic ID counter
}

// NewWorkerPool creates a new worker pool with the specified minimum and maximum number of workers.
func NewWorkerPool[I, O any](ctx context.Context, minWorkers, maxWorkers int, workerFactory workersFactory[I, O]) (WorkerPool[I, O], error) {
	if minWorkers <= 0 {
		return nil, fmt.Errorf("min workers must be greater than zero")
	}

	if maxWorkers < minWorkers {
		return nil, fmt.Errorf("max workers cannot be less than min workers")
	}

	workers := make(workersMap[I, O])
	pool := &workerPool[I, O]{
		ctx:        ctx,
		workers:    workers,
		factory:    workerFactory,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
	}

	for i := 0; i < maxWorkers; i++ {
		pool.addWorker()
	}
	return pool, nil
}

// addWorker adds a single worker to the pool with a unique ID.
func (p *workerPool[I, O]) addWorker() {
	id := atomic.AddUint64(&p.nextID, 1)
	w := p.factory(id)
	w.ID = id
	p.workers[w.ID] = w
	if p.inChan != nil && p.outChan != nil && p.errChan != nil {
		p.StartWorker(w.ID)
	}
}

// StartWorker starts a single worker in the pool.
func (p *workerPool[I, O]) StartWorker(ID uint64) {
	if w, exists := p.workers[ID]; exists {
		w.Start(p.ctx, p.inChan, p.outChan, p.errChan)
	}
}

// Start starts all workers in the pool.
func (p *workerPool[I, O]) Start(inChan chan *I, outChan chan *O, errChan chan error) {
	p.inChan, p.outChan, p.errChan = inChan, outChan, errChan
	for _, w := range p.workers {
		w.Start(p.ctx, p.inChan, p.outChan, p.errChan)
	}
}

// Size returns the current number of workers in the pool.
func (p *workerPool[I, O]) Size() int {
	return len(p.workers)
}

// MinWorkers returns the minimum number of workers in the pool.
func (p *workerPool[I, O]) MinWorkers() int {
	return p.minWorkers
}

// MaxWorkers returns the maximum number of workers in the pool.
func (p *workerPool[I, O]) MaxWorkers() int {
	return p.maxWorkers
}

// AddWorkers adds n workers to the pool.
func (p *workerPool[I, O]) AddWorkers(n int) {
	if n > 0 {
		for i := 0; i < n; i++ {
			p.addWorker()
			worker := p.workers[p.nextID]
			worker.Start(p.ctx, p.inChan, p.outChan, p.errChan)
		}
	}
}

// RemoveWorker removes a worker with the specified ID from the pool.
func (p *workerPool[I, O]) RemoveWorker(id uint64) error {
	if _, exists := p.workers[id]; !exists {
		return fmt.Errorf("worker with ID %d not found", id)
	}
	delete(p.workers, id)
	return nil
}

// RemoveWorkers removes n workers from the pool.
func (p *workerPool[I, O]) RemoveWorkers(n int) error {
	if n <= 0 {
		return fmt.Errorf("number of workers to remove must be greater than zero")
	}
	if n > len(p.workers) {
		return fmt.Errorf("number of workers to remove exceeds current pool size")
	}

	count := 0
	for id := range p.workers {
		delete(p.workers, id)
		count++
		if count == n {
			break
		}
	}
	return nil
}
