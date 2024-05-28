package soop

import (
	"context"
	"fmt"
	"sync/atomic"
)

// WorkerPool interface defines the operations for a pool of workers.
type WorkerPool[I, O any] interface {
	Start()
	Size() int
	MinWorkers() int
	MaxWorkers() int
	AddWorkers(n int)
	GetWorkers() workersMap[I, O]
	RemoveWorker(id uint64) error
	RemoveWorkers(n int) error
	SetWorkerFactory(factory workersFactory[I, O]) error
	SetErrChan(chan error) error
	GetInChan() chan *I
	GetOutChan() chan *O
	GetErrChan() chan error
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
func NewWorkerPool[I, O any](ctx context.Context, minWorkers, maxWorkers int, inChan chan *I, bufferSize int) (WorkerPool[I, O], error) {
	if minWorkers <= 0 {
		return nil, fmt.Errorf("min workers must be greater than zero")
	}

	if maxWorkers < minWorkers {
		return nil, fmt.Errorf("max workers cannot be less than min workers")
	}

	pool := &workerPool[I, O]{
		ctx:        ctx,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		inChan:     inChan,
		outChan:    make(chan *O, bufferSize),
	}

	return pool, nil
}

func (p *workerPool[I, O]) SetWorkerFactory(factory workersFactory[I, O]) error {
	if factory == nil {
		return fmt.Errorf("worker factory cannot be nil")
	}

	p.factory = factory

	p.workers = make(workersMap[I, O])
	for i := 0; i < p.maxWorkers; i++ {
		err := p.addWorker()
		if err != nil {
			return err
		}
	}

	return nil
}

// addWorker adds a single worker to the pool with a unique ID.
func (p *workerPool[I, O]) addWorker() error {
	if p.factory == nil {
		return fmt.Errorf("worker factory not set")
	}
	id := atomic.AddUint64(&p.nextID, 1)
	w := p.factory(id)
	w.ID = id
	p.workers[w.ID] = w
	if p.inChan == nil || p.outChan == nil || p.errChan == nil {
		return nil
	}
	p.StartWorker(w.ID)
	return nil
}

// StartWorker starts a single worker in the pool.
func (p *workerPool[I, O]) StartWorker(ID uint64) {
	if w, exists := p.workers[ID]; exists {
		w.Start()
	}
}

// Start starts all workers in the pool.
func (p *workerPool[I, O]) Start() {
	for _, w := range p.workers {
		w.Start()
	}
}

// GetErrChan returns the error channel of the pool.
func (p *workerPool[I, O]) GetErrChan() chan error {
	return p.errChan
}

// SetErrChan sets the error channel for the pool.
func (p *workerPool[I, O]) SetErrChan(errChan chan error) error {
	if errChan == nil {
		return fmt.Errorf("error channel cannot be nil")
	}
	p.errChan = errChan

	for _, w := range p.workers {
		w.SetErrChan(errChan)
	}
	return nil
}

// GetWorkers returns the map of workers in the pool.
func (p *workerPool[I, O]) GetWorkers() workersMap[I, O] {
	return p.workers
}

// GetInChan returns the input channel of the pool.
func (p *workerPool[I, O]) GetInChan() chan *I {
	return p.inChan
}

// GetOutChan returns the output channel of the pool.
func (p *workerPool[I, O]) GetOutChan() chan *O {
	return p.outChan
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
			worker.Start()
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
