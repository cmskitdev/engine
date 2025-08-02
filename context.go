package engine

import (
	"context"
	"sync"
	"time"
)

// ProcessingContext is a context for the processing engine that is passed to throughout
// the individual processing pipeline process orchestration steps.
type ProcessingContext[T any] struct {
	Context        context.Context
	Request        *PipelineRequest[T]
	Results        chan PipelineResult[T]
	Metrics        *Metrics
	MetricsMu      sync.RWMutex
	MetricsUpdated chan bool
}

// ProcessingContextOptions are options for creating a new processing context.
type ProcessingContextOptions[T any] struct {
	Context    context.Context
	Request    *PipelineRequest[T]
	BufferSize int
	Metrics    bool
}

// NewProcessingContext creates a new processing context.
func NewProcessingContext[T any](opts ProcessingContextOptions[T]) *ProcessingContext[T] {
	var metrics *Metrics
	if opts.Metrics {
		metrics = &Metrics{StartedAt: time.Now()}
	}

	return &ProcessingContext[T]{
		Context:        opts.Context,
		Request:        opts.Request,
		Results:        make(chan PipelineResult[T], opts.BufferSize),
		Metrics:        metrics,
		MetricsMu:      sync.RWMutex{},
		MetricsUpdated: make(chan bool, 1),
	}
}

// WorkerContext is a context for a worker in the processing pool.
type WorkerContext[T any] struct {
	Context context.Context
	ID      int
}
