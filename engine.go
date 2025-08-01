package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProcessingEngine is the core engine that orchestrates data processing
type ProcessingEngine struct {
	config       EngineConfig
	transformers map[string]Transformer
	validators   map[string]Validator
	middleware   []Middleware
	metrics      *EngineMetrics
	mu           sync.RWMutex
}

// EngineConfig configures the processing engine
type EngineConfig struct {
	Name               string        `json:"name"`
	MaxConcurrentItems int           `json:"max_concurrent_items"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableTracing      bool          `json:"enable_tracing"`
	MemoryLimitMB      int           `json:"memory_limit_mb"`
}

// DefaultEngineConfig returns sensible defaults for the engine
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		Name:               "default-engine",
		MaxConcurrentItems: 10,
		DefaultTimeout:     30 * time.Minute,
		EnableMetrics:      true,
		EnableTracing:      false,
		MemoryLimitMB:      512,
	}
}

// EngineMetrics tracks processing metrics
type EngineMetrics struct {
	ItemsProcessed        int64         `json:"items_processed"`
	ItemsSuccessful       int64         `json:"items_successful"`
	ItemsFailed           int64         `json:"items_failed"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	StartedAt             time.Time     `json:"started_at"`
	LastProcessedAt       *time.Time    `json:"last_processed_at,omitempty"`
	mu                    sync.RWMutex
}

// NewProcessingEngine creates a new processing engine
func NewProcessingEngine(config EngineConfig) *ProcessingEngine {
	if config.Name == "" {
		config = DefaultEngineConfig()
	}

	engine := &ProcessingEngine{
		config:       config,
		transformers: make(map[string]Transformer),
		validators:   make(map[string]Validator),
		middleware:   make([]Middleware, 0),
		metrics:      &EngineMetrics{StartedAt: time.Now()},
	}

	return engine
}

// RegisterTransformer registers a transformer with the engine
func (e *ProcessingEngine) RegisterTransformer(name string, transformer Transformer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transformers[name] = transformer
}

// RegisterValidator registers a validator with the engine
func (e *ProcessingEngine) RegisterValidator(name string, validator Validator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.validators[name] = validator
}

// AddMiddleware adds middleware to the processing pipeline
func (e *ProcessingEngine) AddMiddleware(middleware Middleware) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middleware = append(e.middleware, middleware)
}

// Process orchestrates the complete processing of data items
func (e *ProcessingEngine) Process(ctx context.Context, req *PipelineRequest) (<-chan PipelineResult, error) {
	if err := e.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	results := make(chan PipelineResult, req.Options.BufferSize)

	go func() {
		defer close(results)
		e.processItems(ctx, req, results)
	}()

	return results, nil
}

// processItems handles the actual processing workflow hanlded by the `Process` function.
//
// Arguments:
// - ctx: The context for the processing.
// - req: The pipeline request.
// - results: The channel to send the results to.
func (e *ProcessingEngine) processItems(ctx context.Context, req *PipelineRequest, results chan<- PipelineResult) {
	// Phase 1: Read data from source
	// Determine which object types to read based on configuration available in the
	// source (a DataSource object).
	types := e.determineReadTypes(req.Source)

	items, err := req.Source.Read(ctx, &ReadRequest{
		Types:   types,
		Filters: []Filter{}, // Will be populated based on request
		Options: make(map[string]interface{}),
	})
	if err != nil {
		results <- PipelineResult{
			Phase:   PhaseRead,
			Success: false,
			Error:   fmt.Errorf("failed to read from source: %w", err),
		}
		return
	}

	// Create the worker pool for concurrently processing items.
	workerPool := NewWorkerPool(req.Options.MaxWorkers, req.Options.BufferSize)
	defer workerPool.Close()

	// Process items through the pipeline by submitting them to the local worker pool.
	var wg sync.WaitGroup
	for item := range items {
		select {
		// The only time this should happen is if the context is cancelled which is likely
		// due to a timeout or a user cancelling the operation upstream. If this happens,
		// we send a result indicating an error to the results channel.
		case <-ctx.Done():
			results <- PipelineResult{
				Phase:   PhaseError,
				Success: false,
				Error:   ctx.Err(),
			}
			return
		default:
			// Submit the item to the local worker pool for processing.
			wg.Add(1)
			workerPool.Submit(func() {
				defer wg.Done()
				e.processItem(ctx, item, req, results)
			})
		}
	}

	// Block this motherfucker until all items are processed by the current local worker pool.
	wg.Wait()
}

// processItem processes a single data item through all phases
func (e *ProcessingEngine) processItem(ctx context.Context, item DataItemContainer, req *PipelineRequest, results chan<- PipelineResult) {
	startTime := time.Now()
	item.Metadata.ProcessingState.StartedAt = startTime
	item.Metadata.ProcessingState.Status = ProcessingStatusInProgress

	// Apply middleware (pre-processing)
	for _, middleware := range e.middleware {
		if err := middleware.PreProcess(ctx, &item); err != nil {
			e.sendResult(results, item, PhaseError, false, err, startTime)
			return
		}
	}

	// Phase 2: Transform
	if len(req.Transform.Transformers) > 0 {
		item.Metadata.ProcessingState.Phase = PhaseTransform
		transformedItem, err := e.applyTransformations(ctx, item, req.Transform)
		if err != nil {
			e.sendResult(results, item, PhaseTransform, false, err, startTime)
			return
		}
		item = transformedItem
		e.sendResult(results, item, PhaseTransform, true, nil, startTime)
	}

	// Phase 3: Validate
	if !req.Options.SkipValidation && len(req.Validation.Validators) > 0 {
		item.Metadata.ProcessingState.Phase = PhaseValidate
		validationResult, err := e.applyValidations(ctx, item, req.Validation)
		if err != nil {
			e.sendResult(results, item, PhaseValidate, false, err, startTime)
			return
		}
		item.Metadata.ValidationState = validationResult
		e.sendResult(results, item, PhaseValidate, true, nil, startTime)

		// Stop if validation failed and we're in strict mode
		if !validationResult.IsValid && req.Validation.Options.StopOnFirstError {
			e.sendResult(results, item, PhaseError, false, fmt.Errorf("validation failed"), startTime)
			return
		}
	}

	// Phase 4: Write to destination
	if !req.Options.DryRun {
		item.Metadata.ProcessingState.Phase = PhaseWrite
		if err := req.Destination.Write(ctx, item); err != nil {
			e.sendResult(results, item, PhaseWrite, false, err, startTime)
			return
		}
		e.sendResult(results, item, PhaseWrite, true, nil, startTime)
	}

	// Apply middleware (post-processing)
	for _, middleware := range e.middleware {
		if err := middleware.PostProcess(ctx, &item); err != nil {
			e.sendResult(results, item, PhaseError, false, err, startTime)
			return
		}
	}

	// Mark as complete
	item.Metadata.ProcessingState.Phase = PhaseComplete
	item.Metadata.ProcessingState.Status = ProcessingStatusCompleted
	completedAt := time.Now()
	item.Metadata.ProcessingState.CompletedAt = &completedAt

	e.sendResult(results, item, PhaseComplete, true, nil, startTime)
	e.updateMetrics(true, time.Since(startTime))
}

// sendResult sends a processing result
func (e *ProcessingEngine) sendResult(results chan<- PipelineResult, item DataItemContainer, phase ProcessingPhase, success bool, err error, startTime time.Time) {
	results <- PipelineResult{
		Item:    item,
		Phase:   phase,
		Success: success,
		Error:   err,
		Metadata: ResultMetadata{
			ProcessedAt:    time.Now(),
			ProcessingTime: time.Since(startTime),
		},
	}
}

// applyTransformations applies all configured transformations to an item
func (e *ProcessingEngine) applyTransformations(ctx context.Context, item DataItemContainer, config TransformConfig) (DataItemContainer, error) {
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

// applyValidations applies all configured validations to an item
func (e *ProcessingEngine) applyValidations(ctx context.Context, item DataItemContainer, config ValidationConfig) (ValidationState, error) {
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

// validateRequest validates the pipeline request
func (e *ProcessingEngine) validateRequest(req *PipelineRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	if req.Destination == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	// Validate source
	if err := req.Source.Validate(&ReadRequest{}); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	return nil
}

// updateMetrics updates processing metrics
func (e *ProcessingEngine) updateMetrics(success bool, duration time.Duration) {
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
func (e *ProcessingEngine) GetMetrics() EngineMetrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	// Return a copy without the mutex to avoid lock copying
	return EngineMetrics{
		ItemsProcessed:        e.metrics.ItemsProcessed,
		ItemsSuccessful:       e.metrics.ItemsSuccessful,
		ItemsFailed:           e.metrics.ItemsFailed,
		TotalProcessingTime:   e.metrics.TotalProcessingTime,
		AverageProcessingTime: e.metrics.AverageProcessingTime,
		StartedAt:             e.metrics.StartedAt,
		LastProcessedAt:       e.metrics.LastProcessedAt,
	}
}

// determineReadTypes determines which object types to read based on source configuration
func (e *ProcessingEngine) determineReadTypes(source PluginSource) []ObjectType {
	sourceConfig := source.Config()
	var readTypes []ObjectType

	// Check if source is Notion source by examining its configuration
	if sourceConfig.Type == "notion" {
		// Add object types based on what's enabled in the source configuration
		if props := sourceConfig.Properties; props != nil {
			if includeDatabases, ok := props["include_databases"].(bool); ok && includeDatabases {
				readTypes = append(readTypes, ObjectTypeDatabase)
			}
			if includePages, ok := props["include_pages"].(bool); ok && includePages {
				readTypes = append(readTypes, ObjectTypePage)
			}
			if includeBlocks, ok := props["include_blocks"].(bool); ok && includeBlocks {
				readTypes = append(readTypes, ObjectTypeBlock)
			}
			if includeComments, ok := props["include_comments"].(bool); ok && includeComments {
				readTypes = append(readTypes, ObjectTypeComment)
			}
			if includeUsers, ok := props["include_users"].(bool); ok && includeUsers {
				readTypes = append(readTypes, ObjectTypeUser)
			}
		}
	} else {
		// For other source types, include all supported types
		allTypes := []ObjectType{
			ObjectTypeDatabase, ObjectTypePage, ObjectTypeBlock,
			ObjectTypeComment, ObjectTypeUser, ObjectTypeFile,
			ObjectTypeWorkspace, ObjectTypeProperty, ObjectTypeRelation,
		}

		for _, objType := range allTypes {
			if source.SupportsType(objType) {
				readTypes = append(readTypes, objType)
			}
		}
	}

	// If no types were determined, include at least the basic types
	if len(readTypes) == 0 {
		readTypes = []ObjectType{ObjectTypePage, ObjectTypeDatabase}
	}

	return readTypes
}

// Close cleanly shuts down the engine
func (e *ProcessingEngine) Close() error {
	// Close any resources
	return nil
}

// Transformer interface for data transformations
type Transformer interface {
	Transform(ctx context.Context, item DataItemContainer) (DataItemContainer, error)
	GetName() string
	GetDescription() string
}

// Validator interface for data validation
type Validator interface {
	Validate(ctx context.Context, item DataItemContainer) ValidationState
	GetName() string
	GetDescription() string
}

// Middleware interface for pre/post processing
type Middleware interface {
	PreProcess(ctx context.Context, item *DataItemContainer) error
	PostProcess(ctx context.Context, item *DataItemContainer) error
	GetName() string
}
