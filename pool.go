package engine

import (
	"fmt"
	"sync"

	"github.com/mateothegreat/go-multilog/multilog"
)

// WorkerPool manages concurrent processing workers.
type WorkerPool struct {
	workers   int
	workChan  chan func()
	closeChan chan struct{}
	wg        sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers, bufferSize int) *WorkerPool {
	multilog.Debug("engine.pool.NewWorkerPool", "Creating worker pool", map[string]interface{}{
		"workers":    workers,
		"bufferSize": bufferSize,
	})

	pool := &WorkerPool{
		workers:   workers,
		workChan:  make(chan func(), bufferSize),
		closeChan: make(chan struct{}),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

// Submit Adds work to the pool for processing by sending it to the work channel.
// This is usually called by the engine's processItems() function.
//
// Arguments:
// - work: A function to be executed by a worker.
func (p *WorkerPool) Submit(work func()) {
	select {
	case p.workChan <- work:
	case <-p.closeChan:
	}
}

// worker processes work items by receiving them from the work channel and executing them.
//
// Arguments:
// - id: The ID of the worker.
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case work := <-p.workChan:
			multilog.Debug("engine.pool.worker", fmt.Sprintf("[%d] worker received work", id), map[string]interface{}{
				"work": work,
			})
			work()
		case <-p.closeChan:
			return
		}
	}
}

// Close shuts down the worker pool
func (p *WorkerPool) Close() {
	close(p.closeChan)
	p.wg.Wait()
}
