// Package engine package contains the core engine that orchestrates data processing.
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cmskitdev/common"
	"github.com/cmskitdev/plugins"
	"github.com/cmskitdev/redis"
)

// Engine is the core engine that orchestrates data processing.
type Engine[T any] struct {
	config         Config
	configManager  *ConfigManager
	transformers   map[string]Transformer[T]
	validators     map[string]Validator[T]
	middleware     []Middleware[T]
	plugins        map[string]*plugins.Plugin[any]
	metrics        *Metrics
	mu             sync.RWMutex
	bus            *plugins.EventBus[any]
	pluginRegistry *plugins.Registry[any]
}

type NewEngineOptions struct {
	Plugins []string
}

// NewEngine creates a new processing engine.
func NewEngine[T any](opts NewEngineOptions) *Engine[T] {
	config := DefaultConfig()

	bus := plugins.NewEventBus[any]()
	reg := plugins.NewRegistry(bus)

	engine := &Engine[T]{
		config:         config,
		transformers:   make(map[string]Transformer[T]),
		validators:     make(map[string]Validator[T]),
		middleware:     make([]Middleware[T], 0),
		plugins:        make(map[string]*plugins.Plugin[any]),
		metrics:        &Metrics{StartedAt: time.Now()},
		bus:            bus,
		pluginRegistry: reg,
	}

	engine.configManager = NewConfigManager()

	engine.registerAll()

	bus.Publish(plugins.Message{
		Event: plugins.EventMessage,
		Data: redis.DataItem{
			ID:   "test",
			Data: "test",
		},
	})

	return engine
}

// RegisterTransformer registers a transformer with the engine.
func (e *Engine[T]) RegisterTransformer(name string, transformer Transformer[T]) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transformers[name] = transformer
}

// AddMiddleware adds middleware to the processing pipeline.
func (e *Engine[T]) AddMiddleware(middleware Middleware[T]) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middleware = append(e.middleware, middleware)
}

// Start begins the pipeline for the requested `PipelineRequest` configuration.
//
// Arguments:
// - ctx: The context for the processing.
// - req: The pipeline request.
//
// Returns:
// - results: The channel to send the results to.
// - err: The error if the request is invalid.
func (e *Engine[T]) Start(ctx context.Context, req *PipelineRequest[T]) (<-chan PipelineResult[T], error) {
	processingContext := NewProcessingContext(ProcessingContextOptions[T]{
		Context:    ctx,
		Request:    req,
		Metrics:    e.config.EnableMetrics,
		BufferSize: req.Options.BufferSize,
	})

	go func() {
		defer close(processingContext.Results)
		e.work(processingContext)
	}()

	return processingContext.Results, nil
}

// work handles the actual processing workflow hanlded by the `Process` function.
//
// Arguments:
// - ctx: The context for the processing.
// - req: The pipeline request.
// - results: The channel to send the results to.
func (e *Engine[T]) work(pc *ProcessingContext[T]) {
	readReq := &ReadRequest{
		Types: e.determineReadTypes(pc.Request.Source),
		Options: map[string]interface{}{
			"buffer_size": pc.Request.Options.BufferSize,
			"max_workers": pc.Request.Options.MaxWorkers,
		},
	}

	items, err := pc.Request.Source.Read(pc.Context, readReq)
	if err != nil {
		pc.Results <- PipelineResult[T]{Phase: PhaseRead, Success: false, Error: err}
		return
	}

	pool := NewWorkerPool(pc.Request.Options.MaxWorkers, pc.Request.Options.BufferSize)
	defer pool.Close()

	var wg sync.WaitGroup
	for item := range items {
		select {
		case <-pc.Context.Done():
			pc.Results <- PipelineResult[T]{Phase: PhaseError, Success: false, Error: pc.Context.Err()}
			return
		default:
			wg.Add(1)
			pool.Submit(func() {
				defer wg.Done()
				e.transition(pc, item)
			})
		}
	}
	wg.Wait()
}

// transition takes an item and transitions it through all phases in the pipeline
// operating within a single worker goroutine.
//
// Arguments:
// - processingContext: The processing context.
// - item: The item that gets transitioned.
func (e *Engine[T]) transition(processingContext *ProcessingContext[T], item DataItemContainer[T]) {
	startTime := time.Now()
	item.Metadata.ProcessingState.StartedAt = startTime
	item.Metadata.ProcessingState.Status = ProcessingStatusInProgress

	// Apply middleware (pre-processing)
	for _, middleware := range e.middleware {
		if err := middleware.PreProcess(processingContext.Context, &item); err != nil {
			e.sendResult(processingContext.Results, item, PhaseError, false, err, startTime)
			return
		}
	}

	// Phase 2: Transform
	if len(processingContext.Request.Transform.Transformers) > 0 {
		item.Metadata.ProcessingState.Phase = PhaseTransform
		transformedItem, err := e.applyTransformations(processingContext.Context, item, processingContext.Request.Transform)
		if err != nil {
			e.sendResult(processingContext.Results, item, PhaseTransform, false, err, startTime)
			return
		}
		item = transformedItem
		e.sendResult(processingContext.Results, item, PhaseTransform, true, nil, startTime)
	}

	// Phase 3: Validate
	if !processingContext.Request.Options.SkipValidation && len(processingContext.Request.Validation.Validators) > 0 {
		item.Metadata.ProcessingState.Phase = PhaseValidate
		validationResult, err := e.applyValidations(processingContext.Context, item, processingContext.Request.Validation)
		if err != nil {
			e.sendResult(processingContext.Results, item, PhaseValidate, false, err, startTime)
			return
		}
		item.Metadata.ValidationState = validationResult
		e.sendResult(processingContext.Results, item, PhaseValidate, true, nil, startTime)

		// Stop if validation failed and we're in strict mode
		if !validationResult.IsValid && processingContext.Request.Validation.Options.StopOnFirstError {
			e.sendResult(processingContext.Results, item, PhaseError, false, fmt.Errorf("validation failed"), startTime)
			return
		}
	}

	// Phase 4: Write to destination
	if !processingContext.Request.Options.DryRun {
		item.Metadata.ProcessingState.Phase = PhaseWrite
		if err := processingContext.Request.Destination.Write(processingContext.Context, item); err != nil {
			e.sendResult(processingContext.Results, item, PhaseWrite, false, err, startTime)
			return
		}
		e.sendResult(processingContext.Results, item, PhaseWrite, true, nil, startTime)
	}

	// Apply middleware (post-processing)
	for _, middleware := range e.middleware {
		if err := middleware.PostProcess(processingContext.Context, &item); err != nil {
			e.sendResult(processingContext.Results, item, PhaseError, false, err, startTime)
			return
		}
	}

	// Mark as complete
	item.Metadata.ProcessingState.Phase = PhaseComplete
	item.Metadata.ProcessingState.Status = ProcessingStatusCompleted
	completedAt := time.Now()
	item.Metadata.ProcessingState.CompletedAt = &completedAt

	e.sendResult(processingContext.Results, item, PhaseComplete, true, nil, startTime)
	e.metric(true, time.Since(startTime))
}

// sendResult sends a processing result.
func (e *Engine[T]) sendResult(results chan<- PipelineResult[T], item DataItemContainer[T], phase ProcessingPhase, success bool, err error, startTime time.Time) {
	results <- PipelineResult[T]{
		Item:    item,
		Phase:   phase,
		Success: success,
		Error:   err,
		Metadata: ResultMetadata{
			Time:     time.Now(),
			Duration: time.Since(startTime),
		},
	}
}

// applyTransformations applies all configured transformations to an item.
func (e *Engine[T]) applyTransformations(ctx context.Context, item DataItemContainer[T], config TransformConfig) (DataItemContainer[T], error) {
	transformedItem := item
	transformLog := make([]TransformStep, 0, len(config.Transformers))

	for _, transformerConfig := range config.Transformers {
		stepStart := time.Now()

		e.mu.RLock()
		transformer, exists := e.transformers[transformerConfig.Type]
		e.mu.RUnlock()

		if !exists {
			step := TransformStep{
				StepName: transformerConfig.Type,
				Applied:  false,
				Error:    fmt.Sprintf("transformer '%s' not found", transformerConfig.Type),
				Duration: time.Since(stepStart),
			}
			transformLog = append(transformLog, step)

			if !config.Options.SkipOnError {
				transformedItem.Metadata.TransformState.TransformLog = transformLog
				return transformedItem, fmt.Errorf("transformer '%s' not found", transformerConfig.Type)
			}
			continue
		}

		result, err := transformer.Transform(ctx, transformedItem)
		step := TransformStep{
			StepName: transformerConfig.Type,
			Applied:  err == nil,
			Duration: time.Since(stepStart),
		}

		if err != nil {
			step.Error = err.Error()
			transformLog = append(transformLog, step)

			if !config.Options.SkipOnError {
				transformedItem.Metadata.TransformState.TransformLog = transformLog
				return transformedItem, fmt.Errorf("transformation failed at step '%s': %w", transformerConfig.Type, err)
			}
		} else {
			transformedItem = result
			transformLog = append(transformLog, step)
		}
	}

	// Update transform state
	transformedItem.Metadata.TransformState = TransformState{
		IsTransformed: true,
		TransformedAt: time.Now(),
		TransformLog:  transformLog,
	}

	return transformedItem, nil
}

// applyValidations applies all configured validations to an item.
func (e *Engine[T]) applyValidations(ctx context.Context, item DataItemContainer[T], config ValidationConfig) (ValidationState, error) {
	state := ValidationState{
		IsValid:     true,
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]ValidationWarning, 0),
		ValidatedAt: time.Now(),
		Rules:       make([]string, 0, len(config.Validators)),
	}

	for _, validatorConfig := range config.Validators {
		e.mu.RLock()
		validator, exists := e.validators[validatorConfig.Type]
		e.mu.RUnlock()

		if !exists {
			state.Errors = append(state.Errors, ValidationError{
				Code:     "VALIDATOR_NOT_FOUND",
				Message:  fmt.Sprintf("validator '%s' not found", validatorConfig.Type),
				Severity: ValidationSeverityError,
			})
			state.IsValid = false
			continue
		}

		result := validator.Validate(ctx, item)
		state.Rules = append(state.Rules, validatorConfig.Type)

		// Merge results
		state.Errors = append(state.Errors, result.Errors...)
		state.Warnings = append(state.Warnings, result.Warnings...)

		if !result.IsValid {
			state.IsValid = false
		}

		// Stop on first error if configured
		if !result.IsValid && config.Options.StopOnFirstError {
			break
		}
	}

	return state, nil
}

// metric updates processing metrics.
func (e *Engine[T]) metric(success bool, duration time.Duration) {
	if !e.config.EnableMetrics {
		return
	}

	e.metrics.mu.Lock()
	defer e.metrics.mu.Unlock()

	e.metrics.ItemsProcessed++
	if success {
		e.metrics.ItemsSuccessful++
	} else {
		e.metrics.ItemsFailed++
	}

	e.metrics.TotalProcessingTime += duration
	e.metrics.AverageProcessingTime = time.Duration(int64(e.metrics.TotalProcessingTime) / e.metrics.ItemsProcessed)

	now := time.Now()
	e.metrics.LastProcessedAt = &now
}

// GetMetrics returns current processing metrics (copy without mutex)
func (e *Engine[T]) GetMetrics() Metrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	// Return a copy without the mutex to avoid lock copying.
	return Metrics{
		ItemsProcessed:        e.metrics.ItemsProcessed,
		ItemsSuccessful:       e.metrics.ItemsSuccessful,
		ItemsFailed:           e.metrics.ItemsFailed,
		TotalProcessingTime:   e.metrics.TotalProcessingTime,
		AverageProcessingTime: e.metrics.AverageProcessingTime,
		StartedAt:             e.metrics.StartedAt,
		LastProcessedAt:       e.metrics.LastProcessedAt,
	}
}

// determineReadTypes determines which object types to read based on source configuration.
func (e *Engine[T]) determineReadTypes(source PluginSource[T]) []common.ObjectType {
	sourceConfig := source.Config()
	var readTypes []common.ObjectType

	// Check if source is Notion source by examining its configuration.
	if sourceConfig.Type == "notion" {
		// Add object types based on what's enabled in the source configuration.
		if props := sourceConfig.Properties; props != nil {
			if includeDatabases, ok := props["include_databases"].(bool); ok && includeDatabases {
				readTypes = append(readTypes, common.ObjectTypeCollection)
			}
			if includePages, ok := props["include_pages"].(bool); ok && includePages {
				readTypes = append(readTypes, common.ObjectTypePage)
			}
			if includeBlocks, ok := props["include_blocks"].(bool); ok && includeBlocks {
				readTypes = append(readTypes, common.ObjectTypeBlock)
			}
			if includeComments, ok := props["include_comments"].(bool); ok && includeComments {
				readTypes = append(readTypes, common.ObjectTypeComment)
			}
			if includeUsers, ok := props["include_users"].(bool); ok && includeUsers {
				readTypes = append(readTypes, common.ObjectTypeUser)
			}
		}
	} else {
		// For other source types, include all supported types
		for _, objType := range common.ObjectTypes {
			if source.SupportsType(objType) {
				readTypes = append(readTypes, objType)
			}
		}
	}

	// If no types were determined, include at least the basic types
	if len(readTypes) == 0 {
		readTypes = []common.ObjectType{common.ObjectTypePage, common.ObjectTypeCollection}
	}

	return readTypes
}

// Close cleanly shuts down the engine
func (e *Engine[T]) Close() error {
	// Close any resources
	return nil
}

// Transformer interface for data transformations
type Transformer[T any] interface {
	Transform(ctx context.Context, item DataItemContainer[T]) (DataItemContainer[T], error)
	GetName() string
	GetDescription() string
}

// Validator interface for data validation
type Validator[T any] interface {
	Validate(ctx context.Context, item DataItemContainer[T]) ValidationState
	GetName() string
	GetDescription() string
}

// Middleware interface for pre/post processing
type Middleware[T any] interface {
	PreProcess(ctx context.Context, item *DataItemContainer[T]) error
	PostProcess(ctx context.Context, item *DataItemContainer[T]) error
	GetName() string
}
