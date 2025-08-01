package engine

import (
	"context"
	"fmt"
	"time"
)

// ExportOperation handles data export from sources to destinations
type ExportOperation struct {
	engine *ProcessingEngine
	config ExportConfig
}

// ExportConfig configures export operations
type ExportConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// Processing options
	ConcurrentWorkers int           `json:"concurrent_workers"`
	BatchSize         int           `json:"batch_size"`
	Timeout           time.Duration `json:"timeout"`

	// Error handling
	ContinueOnError bool `json:"continue_on_error"`
	MaxErrors       int  `json:"max_errors"`

	// Progress tracking
	ReportProgress   bool          `json:"report_progress"`
	ProgressInterval time.Duration `json:"progress_interval"`

	// Output options
	DryRun         bool `json:"dry_run"`
	Compress       bool `json:"compress"`
	IncludeDeleted bool `json:"include_deleted"`
}

// DefaultExportConfig returns sensible defaults for export operations
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		Name:              "default-export",
		Description:       "Default export operation",
		ConcurrentWorkers: 5,
		BatchSize:         100,
		Timeout:           30 * time.Minute,
		ContinueOnError:   true,
		MaxErrors:         10,
		ReportProgress:    true,
		ProgressInterval:  10 * time.Second,
		DryRun:            false,
		Compress:          false,
		IncludeDeleted:    false,
	}
}

// NewExportOperation creates a new export operation
func NewExportOperation(engine *ProcessingEngine, config ExportConfig) *ExportOperation {
	return &ExportOperation{
		engine: engine,
		config: config,
	}
}

// ExportRequest defines what and how to export
type ExportRequest struct {
	// Source configuration
	Source PluginSource `json:"source"`

	// Destination configuration
	Destination PluginDestination `json:"destination"`

	// What to export
	Scope ExportScope `json:"scope"`

	// Transformation pipeline
	Transforms []TransformerConfig `json:"transforms,omitempty"`

	// Validation rules
	Validation ValidationConfig `json:"validation,omitempty"`

	// Export-specific options
	Options ExportOptions `json:"options"`
}

// ExportScope defines what data to export
type ExportScope struct {
	// Object types to include
	ObjectTypes []ObjectType `json:"object_types,omitempty"`

	// Specific object IDs to export
	ObjectIDs []string `json:"object_ids,omitempty"`

	// Filters to apply
	Filters []Filter `json:"filters,omitempty"`

	// Date range filters
	CreatedAfter   *time.Time `json:"created_after,omitempty"`
	CreatedBefore  *time.Time `json:"created_before,omitempty"`
	ModifiedAfter  *time.Time `json:"modified_after,omitempty"`
	ModifiedBefore *time.Time `json:"modified_before,omitempty"`

	// Hierarchical options
	IncludeChildren  bool `json:"include_children"`
	IncludeParents   bool `json:"include_parents"`
	MaxDepth         int  `json:"max_depth"`         // 0 = no limit
	FollowReferences bool `json:"follow_references"` // Follow relationships
}

// ExportOptions provides additional export configuration
type ExportOptions struct {
	// Data inclusion options
	IncludeMetadata      bool `json:"include_metadata"`
	IncludeRelationships bool `json:"include_relationships"`
	IncludeDependencies  bool `json:"include_dependencies"`

	// Output format options
	CompressOutput  bool   `json:"compress_output"`
	SplitLargeFiles bool   `json:"split_large_files"`
	MaxFileSize     int64  `json:"max_file_size"` // in bytes
	OutputFormat    string `json:"output_format"` // json, csv, markdown, etc.

	// Performance options
	StreamingMode    bool `json:"streaming_mode"`    // Stream results vs batch
	MemoryOptimized  bool `json:"memory_optimized"`  // Optimize for low memory usage
	NetworkOptimized bool `json:"network_optimized"` // Optimize for network efficiency

	// Resume options
	Resumable      bool   `json:"resumable"`       // Support resume on failure
	CheckpointFile string `json:"checkpoint_file"` // File to store progress
}

// ExportResult contains the results of an export operation
type ExportResult struct {
	// Operation metadata
	OperationID string        `json:"operation_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`

	// Statistics
	Stats ExportStats `json:"stats"`

	// Processing summary
	Success  bool     `json:"success"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`

	// Output information
	OutputLocation string                 `json:"output_location,omitempty"`
	OutputSize     int64                  `json:"output_size,omitempty"`
	OutputFormat   string                 `json:"output_format,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ExportStats contains detailed statistics about the export operation
type ExportStats struct {
	// Item counts by type
	ItemsProcessed map[ObjectType]int64 `json:"items_processed"`
	ItemsSucceeded map[ObjectType]int64 `json:"items_succeeded"`
	ItemsFailed    map[ObjectType]int64 `json:"items_failed"`
	ItemsSkipped   map[ObjectType]int64 `json:"items_skipped"`

	// Totals
	TotalItems      int64 `json:"total_items"`
	SuccessfulItems int64 `json:"successful_items"`
	FailedItems     int64 `json:"failed_items"`
	SkippedItems    int64 `json:"skipped_items"`

	// Performance metrics
	ItemsPerSecond  float64       `json:"items_per_second"`
	BytesPerSecond  float64       `json:"bytes_per_second"`
	AverageItemTime time.Duration `json:"average_item_time"`
	NetworkRequests int64         `json:"network_requests"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`

	// Resource usage
	MaxMemoryUsage       int64 `json:"max_memory_usage"` // in bytes
	TotalBytesRead       int64 `json:"total_bytes_read"`
	TotalBytesWritten    int64 `json:"total_bytes_written"`
	NetworkBytesReceived int64 `json:"network_bytes_received"`
	NetworkBytesSent     int64 `json:"network_bytes_sent"`
}

// Execute performs the export operation
func (op *ExportOperation) Execute(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &ExportResult{
		OperationID: fmt.Sprintf("export-%d", startTime.Unix()),
		StartTime:   startTime,
		Stats: ExportStats{
			ItemsProcessed: make(map[ObjectType]int64),
			ItemsSucceeded: make(map[ObjectType]int64),
			ItemsFailed:    make(map[ObjectType]int64),
			ItemsSkipped:   make(map[ObjectType]int64),
		},
		Metadata: make(map[string]interface{}),
	}

	// Validate request
	if err := op.validateRequest(req); err != nil {
		result.Success = false
		result.Errors = []string{fmt.Sprintf("validation failed: %v", err)}
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Create pipeline request
	pipelineReq := &PipelineRequest{
		Direction:   DirectionExport,
		Source:      req.Source,
		Destination: req.Destination,
		Transform: TransformConfig{
			Transformers: req.Transforms,
			Options: TransformOptions{
				SkipOnError:      req.Options.StreamingMode,
				PreserveMetadata: req.Options.IncludeMetadata,
			},
		},
		Validation: req.Validation,
		Options: PipelineOptions{
			MaxWorkers:       op.config.ConcurrentWorkers,
			BatchSize:        op.config.BatchSize,
			BufferSize:       op.config.BatchSize * 2,
			Timeout:          op.config.Timeout,
			ContinueOnError:  op.config.ContinueOnError,
			MaxErrors:        op.config.MaxErrors,
			EnableProgress:   op.config.ReportProgress,
			ProgressInterval: op.config.ProgressInterval,
			DryRun:           op.config.DryRun,
		},
	}

	// Process the pipeline
	results, err := op.engine.Process(ctx, pipelineReq)
	if err != nil {
		result.Success = false
		result.Errors = []string{fmt.Sprintf("pipeline failed: %v", err)}
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Process results and collect statistics
	var lastError error
	errorCount := 0

	for pipelineResult := range results {
		// Update statistics
		op.updateStats(&result.Stats, pipelineResult)

		// Handle errors
		if pipelineResult.Error != nil {
			lastError = pipelineResult.Error
			errorCount++

			if errorCount > op.config.MaxErrors && !op.config.ContinueOnError {
				result.Success = false
				result.Errors = append(result.Errors, fmt.Sprintf("too many errors, stopping: %v", pipelineResult.Error))
				break
			}
		}

		// Report progress if enabled
		if op.config.ReportProgress && result.Stats.TotalItems%int64(op.config.BatchSize) == 0 {
			fmt.Printf("Exported %d items...\n", result.Stats.TotalItems)
		}
	}

	// Finalize results
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)
	result.Success = errorCount == 0 || (op.config.ContinueOnError && errorCount <= op.config.MaxErrors)

	// Calculate performance metrics
	op.calculatePerformanceMetrics(&result.Stats, result.Duration)

	// Add metadata
	result.Metadata["source_type"] = req.Source.Config().Type
	result.Metadata["destination_type"] = req.Destination.Config().Type
	result.Metadata["scope_object_types"] = req.Scope.ObjectTypes
	result.Metadata["scope_include_children"] = req.Scope.IncludeChildren
	result.Metadata["transforms_count"] = len(req.Transforms)

	if lastError != nil && !result.Success {
		return result, lastError
	}

	return result, nil
}

// validateRequest validates the export request
func (op *ExportOperation) validateRequest(req *ExportRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	if req.Destination == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	// Validate scope
	if len(req.Scope.ObjectTypes) == 0 && len(req.Scope.ObjectIDs) == 0 {
		return fmt.Errorf("scope must specify either object types or object IDs")
	}

	// Validate that source supports the requested object types
	for _, objType := range req.Scope.ObjectTypes {
		if !req.Source.SupportsType(objType) {
			return fmt.Errorf("source does not support object type: %s", objType)
		}
	}

	// Validate that destination supports the requested object types
	for _, objType := range req.Scope.ObjectTypes {
		if !req.Destination.SupportsType(objType) {
			return fmt.Errorf("destination does not support object type: %s", objType)
		}
	}

	// Validate date ranges
	if req.Scope.CreatedAfter != nil && req.Scope.CreatedBefore != nil {
		if req.Scope.CreatedAfter.After(*req.Scope.CreatedBefore) {
			return fmt.Errorf("created_after cannot be after created_before")
		}
	}

	if req.Scope.ModifiedAfter != nil && req.Scope.ModifiedBefore != nil {
		if req.Scope.ModifiedAfter.After(*req.Scope.ModifiedBefore) {
			return fmt.Errorf("modified_after cannot be after modified_before")
		}
	}

	return nil
}

// updateStats updates statistics based on pipeline results
func (op *ExportOperation) updateStats(stats *ExportStats, result PipelineResult) {
	stats.TotalItems++

	if stats.ItemsProcessed[result.Item.Type] == 0 {
		stats.ItemsProcessed[result.Item.Type] = 0
		stats.ItemsSucceeded[result.Item.Type] = 0
		stats.ItemsFailed[result.Item.Type] = 0
		stats.ItemsSkipped[result.Item.Type] = 0
	}

	stats.ItemsProcessed[result.Item.Type]++

	if result.Success {
		stats.ItemsSucceeded[result.Item.Type]++
		stats.SuccessfulItems++
	} else {
		stats.ItemsFailed[result.Item.Type]++
		stats.FailedItems++
	}

	// Update bytes processed if available
	if sizeBytes, exists := result.Item.Metadata.Properties["size_bytes"]; exists {
		if size, ok := sizeBytes.(int64); ok {
			stats.TotalBytesRead += size
		}
	}
}

// calculatePerformanceMetrics calculates performance metrics
func (op *ExportOperation) calculatePerformanceMetrics(stats *ExportStats, duration time.Duration) {
	if duration.Seconds() > 0 {
		stats.ItemsPerSecond = float64(stats.TotalItems) / duration.Seconds()
		stats.BytesPerSecond = float64(stats.TotalBytesRead) / duration.Seconds()
	}

	if stats.TotalItems > 0 {
		stats.AverageItemTime = duration / time.Duration(stats.TotalItems)
	}
}

// ExportBuilder provides a fluent interface for building export operations
type ExportBuilder struct {
	request *ExportRequest
	config  ExportConfig
}

// NewExportBuilder creates a new export builder
func NewExportBuilder() *ExportBuilder {
	return &ExportBuilder{
		request: &ExportRequest{
			Scope:   ExportScope{},
			Options: ExportOptions{},
		},
		config: DefaultExportConfig(),
	}
}

// From sets the data source
func (b *ExportBuilder) From(source PluginSource) *ExportBuilder {
	b.request.Source = source
	return b
}

// To sets the data destination
func (b *ExportBuilder) To(destination PluginDestination) *ExportBuilder {
	b.request.Destination = destination
	return b
}

// Include adds object types to the export scope
func (b *ExportBuilder) Include(objectTypes ...ObjectType) *ExportBuilder {
	b.request.Scope.ObjectTypes = append(b.request.Scope.ObjectTypes, objectTypes...)
	return b
}

// WithIDs sets specific object IDs to export
func (b *ExportBuilder) WithIDs(ids ...string) *ExportBuilder {
	b.request.Scope.ObjectIDs = append(b.request.Scope.ObjectIDs, ids...)
	return b
}

// WithChildren includes child objects in the export
func (b *ExportBuilder) WithChildren(maxDepth int) *ExportBuilder {
	b.request.Scope.IncludeChildren = true
	b.request.Scope.MaxDepth = maxDepth
	return b
}

// WithFilter adds a filter to the export scope
func (b *ExportBuilder) WithFilter(field string, op FilterOp, value interface{}) *ExportBuilder {
	filter := Filter{
		Field:     field,
		Operation: op,
		Value:     value,
	}
	b.request.Scope.Filters = append(b.request.Scope.Filters, filter)
	return b
}

// CreatedAfter sets a created after filter
func (b *ExportBuilder) CreatedAfter(t time.Time) *ExportBuilder {
	b.request.Scope.CreatedAfter = &t
	return b
}

// ModifiedAfter sets a modified after filter
func (b *ExportBuilder) ModifiedAfter(t time.Time) *ExportBuilder {
	b.request.Scope.ModifiedAfter = &t
	return b
}

// Transform adds a transformation to the pipeline
func (b *ExportBuilder) Transform(transformerType string, properties map[string]interface{}) *ExportBuilder {
	transform := TransformerConfig{
		Type:       transformerType,
		Properties: properties,
	}
	b.request.Transforms = append(b.request.Transforms, transform)
	return b
}

// WithFormat sets the output format
func (b *ExportBuilder) WithFormat(format string) *ExportBuilder {
	b.request.Options.OutputFormat = format
	return b
}

// Compress enables output compression
func (b *ExportBuilder) Compress() *ExportBuilder {
	b.request.Options.CompressOutput = true
	return b
}

// DryRun enables dry run mode
func (b *ExportBuilder) DryRun() *ExportBuilder {
	b.config.DryRun = true
	return b
}

// WithConcurrency sets the number of concurrent workers
func (b *ExportBuilder) WithConcurrency(workers int) *ExportBuilder {
	b.config.ConcurrentWorkers = workers
	return b
}

// Build creates the export operation
func (b *ExportBuilder) Build(engine *ProcessingEngine) *ExportOperation {
	return NewExportOperation(engine, b.config)
}

// Execute builds and executes the export operation
func (b *ExportBuilder) Execute(ctx context.Context, engine *ProcessingEngine) (*ExportResult, error) {
	operation := b.Build(engine)
	return operation.Execute(ctx, b.request)
}
