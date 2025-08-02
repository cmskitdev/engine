// package engine provides the core engine for processing data items.
package engine

import (
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

// NewWorkerPool creates a new worker pool.
//
// Arguments:
// - workers: The number of workers to create.
// - bufferSize: The size of the buffer for the work channel.
//
// Returns:
// - A pointer to the new worker pool.
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

		workerContext := &WorkerContext[any]{
			ID: i,
		}

		go pool.worker(workerContext)
	}

	return pool
}

// Submit Adds work to the pool for processing by sending it to the work channel.
// This is usually called by the engine's processItems() function.
//
// Arguments:
// - work: Function to be executed by a worker.
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
func (p *WorkerPool) worker(workerContext *WorkerContext[any]) {
	defer p.wg.Done()
	for {
		select {
		case work := <-p.workChan:
			multilog.Trace("engine.pool.worker", "doing work()", map[string]interface{}{
				"worker": workerContext.ID,
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
	multilog.Debug("engine.pool.Close", "Shutting down worker pool", map[string]interface{}{
		"workers": p.workers,
	})
	p.wg.Wait()
}
