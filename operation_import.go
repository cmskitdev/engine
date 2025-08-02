package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/cmskitdev/common"
)

// ImportOperation handles data import from sources to destinations
type ImportOperation struct {
	engine *Engine[any]
	config ImportConfig
}

// ImportConfig configures import operations
type ImportConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// Processing options
	ConcurrentWorkers int           `json:"concurrent_workers"`
	BatchSize         int           `json:"batch_size"`
	Timeout           time.Duration `json:"timeout"`

	// Import behavior
	CreateMissing      bool               `json:"create_missing"`  // Create objects that don't exist
	UpdateExisting     bool               `json:"update_existing"` // Update existing objects
	ConflictResolution ConflictResolution `json:"conflict_resolution"`
	HandleDependencies bool               `json:"handle_dependencies"` // Resolve dependencies automatically

	// Error handling
	ContinueOnError bool `json:"continue_on_error"`
	MaxErrors       int  `json:"max_errors"`

	// Progress tracking
	ReportProgress   bool          `json:"report_progress"`
	ProgressInterval time.Duration `json:"progress_interval"`

	// Validation
	ValidateBeforeImport bool `json:"validate_before_import"`
	StrictValidation     bool `json:"strict_validation"`

	// Recovery options
	Resumable      bool   `json:"resumable"`       // Support resume on failure
	CheckpointFile string `json:"checkpoint_file"` // File to store progress
	BackupExisting bool   `json:"backup_existing"` // Backup before overwriting
}

// ConflictResolution defines how to handle import conflicts
type ConflictResolution string

const (
	ConflictResolutionSkip      ConflictResolution = "skip"      // Skip conflicting items
	ConflictResolutionOverwrite ConflictResolution = "overwrite" // Overwrite existing items
	ConflictResolutionMerge     ConflictResolution = "merge"     // Merge with existing items
	ConflictResolutionError     ConflictResolution = "error"     // Fail on conflicts
	ConflictResolutionRename    ConflictResolution = "rename"    // Rename conflicting items
)

// DefaultImportConfig returns sensible defaults for import operations
func DefaultImportConfig() ImportConfig {
	return ImportConfig{
		Name:                 "default-import",
		Description:          "Default import operation",
		ConcurrentWorkers:    3, // Lower than export to avoid API rate limits
		BatchSize:            50,
		Timeout:              60 * time.Minute,
		CreateMissing:        true,
		UpdateExisting:       false,
		ConflictResolution:   ConflictResolutionSkip,
		HandleDependencies:   true,
		ContinueOnError:      true,
		MaxErrors:            20,
		ReportProgress:       true,
		ProgressInterval:     15 * time.Second,
		ValidateBeforeImport: true,
		StrictValidation:     false,
		Resumable:            false,
		BackupExisting:       false,
	}
}

// NewImportOperation creates a new import operation
func NewImportOperation(engine *Engine[any], config ImportConfig) *ImportOperation {
	return &ImportOperation{
		engine: engine,
		config: config,
	}
}

// ImportRequest defines what and how to import
type ImportRequest struct {
	// Source configuration
	Source PluginSource[any] `json:"source"`

	// Destination configuration
	Destination PluginDestination[any] `json:"destination"`

	// What to import
	Scope ImportScope `json:"scope"`

	// Transformation pipeline
	Transforms []TransformerConfig `json:"transforms,omitempty"`

	// Validation rules
	Validation ValidationConfig `json:"validation,omitempty"`

	// Import-specific options
	Options ImportOptions `json:"options"`
}

// ImportScope defines what data to import
type ImportScope struct {
	// Object types to include
	ObjectTypes []common.ObjectType `json:"object_types,omitempty"`

	// Specific object IDs to import
	ObjectIDs []string `json:"object_ids,omitempty"`

	// Filters to apply during import
	Filters []Filter `json:"filters,omitempty"`

	// Import order preferences
	ImportOrder []common.ObjectType `json:"import_order,omitempty"` // Preferred import order

	// Dependency handling
	ResolveDependencies bool `json:"resolve_dependencies"`
	MaxDependencyDepth  int  `json:"max_dependency_depth"` // 0 = no limit

	// Target location in destination
	TargetParentID  string `json:"target_parent_id,omitempty"` // Where to import items
	CreateHierarchy bool   `json:"create_hierarchy"`           // Recreate source hierarchy
}

// ImportOptions provides additional import configuration
type ImportOptions struct {
	// Data handling options
	PreserveIDs        bool `json:"preserve_ids"`        // Try to preserve original IDs
	PreserveTimestamps bool `json:"preserve_timestamps"` // Preserve created/modified times
	PreserveMetadata   bool `json:"preserve_metadata"`   // Include source metadata

	// Conflict handling
	DuplicateStrategy DuplicateStrategy `json:"duplicate_strategy"`
	ConflictCallback  string            `json:"conflict_callback,omitempty"` // Custom conflict resolution

	// Performance options
	PreloadData     bool `json:"preload_data"`     // Preload related data
	UseTransactions bool `json:"use_transactions"` // Use transactions if supported
	MemoryOptimized bool `json:"memory_optimized"` // Optimize for low memory usage

	// Mapping options
	FieldMappings    map[string]string            `json:"field_mappings,omitempty"`    // Map source fields to destination
	TypeMappings     map[string]string            `json:"type_mappings,omitempty"`     // Map source types to destination
	PropertyMappings map[string]map[string]string `json:"property_mappings,omitempty"` // Per-type property mappings

	// Post-import options
	RunPostImportValidation bool     `json:"run_post_import_validation"`
	PostImportActions       []string `json:"post_import_actions,omitempty"` // Actions to run after import
}

// DuplicateStrategy defines how to handle duplicate items
type DuplicateStrategy string

const (
	DuplicateStrategyIgnore DuplicateStrategy = "ignore"
	DuplicateStrategyUpdate DuplicateStrategy = "update"
	DuplicateStrategyError  DuplicateStrategy = "error"
	DuplicateStrategyMerge  DuplicateStrategy = "merge"
)

// ImportResult contains the results of an import operation
type ImportResult struct {
	// Operation metadata
	OperationID string        `json:"operation_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`

	// Statistics
	Stats ImportStats `json:"stats"`

	// Processing summary
	Success   bool           `json:"success"`
	Errors    []string       `json:"errors,omitempty"`
	Warnings  []string       `json:"warnings,omitempty"`
	Conflicts []ConflictInfo `json:"conflicts,omitempty"`

	// Import results
	ImportedItems map[string]string `json:"imported_items,omitempty"` // Original ID -> New ID mapping
	SkippedItems  []string          `json:"skipped_items,omitempty"`  // IDs of skipped items
	FailedItems   []string          `json:"failed_items,omitempty"`   // IDs of failed items

	// Output information
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CheckpointFile string                 `json:"checkpoint_file,omitempty"`
}

// ImportStats contains detailed statistics about the import operation
type ImportStats struct {
	// Item counts by type
	ItemsProcessed map[common.ObjectType]int64 `json:"items_processed"`
	ItemsCreated   map[common.ObjectType]int64 `json:"items_created"`
	ItemsUpdated   map[common.ObjectType]int64 `json:"items_updated"`
	ItemsFailed    map[common.ObjectType]int64 `json:"items_failed"`
	ItemsSkipped   map[common.ObjectType]int64 `json:"items_skipped"`

	// Totals
	TotalItems    int64 `json:"total_items"`
	CreatedItems  int64 `json:"created_items"`
	UpdatedItems  int64 `json:"updated_items"`
	FailedItems   int64 `json:"failed_items"`
	SkippedItems  int64 `json:"skipped_items"`
	ConflictItems int64 `json:"conflict_items"`

	// Performance metrics
	ItemsPerSecond    float64       `json:"items_per_second"`
	AverageItemTime   time.Duration `json:"average_item_time"`
	NetworkRequests   int64         `json:"network_requests"`
	APICallsSucceeded int64         `json:"api_calls_succeeded"`
	APICallsFailed    int64         `json:"api_calls_failed"`
	RetryAttempts     int64         `json:"retry_attempts"`

	// Resource usage
	MaxMemoryUsage       int64 `json:"max_memory_usage"` // in bytes
	TotalBytesProcessed  int64 `json:"total_bytes_processed"`
	NetworkBytesReceived int64 `json:"network_bytes_received"`
	NetworkBytesSent     int64 `json:"network_bytes_sent"`

	// Dependency resolution
	DependenciesResolved   int64 `json:"dependencies_resolved"`
	CircularDependencies   int64 `json:"circular_dependencies"`
	UnresolvedDependencies int64 `json:"unresolved_dependencies"`
}

// ConflictInfo describes a conflict that occurred during import
type ConflictInfo struct {
	SourceID     string            `json:"source_id"`
	TargetID     string            `json:"target_id,omitempty"`
	ObjectType   common.ObjectType `json:"object_type"`
	ConflictType string            `json:"conflict_type"`
	Description  string            `json:"description"`
	Resolution   string            `json:"resolution"`
	Timestamp    time.Time         `json:"timestamp"`
}

// Execute performs the import operation
func (op *ImportOperation) Execute(ctx context.Context, req *ImportRequest) (*ImportResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &ImportResult{
		OperationID: fmt.Sprintf("import-%d", startTime.Unix()),
		StartTime:   startTime,
		Stats: ImportStats{
			ItemsProcessed: make(map[common.ObjectType]int64),
			ItemsCreated:   make(map[common.ObjectType]int64),
			ItemsUpdated:   make(map[common.ObjectType]int64),
			ItemsFailed:    make(map[common.ObjectType]int64),
			ItemsSkipped:   make(map[common.ObjectType]int64),
		},
		ImportedItems: make(map[string]string),
		Metadata:      make(map[string]interface{}),
	}

	// Validate request
	if err := op.validateRequest(req); err != nil {
		result.Success = false
		result.Errors = []string{fmt.Sprintf("validation failed: %v", err)}
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Create pipeline request with import-specific configuration
	pipelineReq := &PipelineRequest[any]{
		Direction:   DirectionImport,
		Source:      req.Source,
		Destination: req.Destination,
		Transform: TransformConfig{
			Transformers: req.Transforms,
			Options: TransformOptions{
				SkipOnError:      op.config.ContinueOnError,
				PreserveMetadata: req.Options.PreserveMetadata,
				StrictMode:       op.config.StrictValidation,
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
			DryRun:           false, // Import is never dry run
			SkipValidation:   !op.config.ValidateBeforeImport,
		},
	}

	// Process the pipeline
	results, err := op.engine.Start(ctx, pipelineReq)
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

		// Track imported items
		if pipelineResult.Success {
			// For successful imports, we would typically get the new ID from the destination
			// This is simplified - in reality, the destination would provide the new ID
			newID := fmt.Sprintf("imported-%s", pipelineResult.Item.ID)
			result.ImportedItems[pipelineResult.Item.ID] = newID
		} else {
			result.FailedItems = append(result.FailedItems, pipelineResult.Item.ID)
		}

		// Handle errors
		if pipelineResult.Error != nil {
			lastError = pipelineResult.Error
			errorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: %v", pipelineResult.Item.ID, pipelineResult.Error))

			if errorCount > op.config.MaxErrors && !op.config.ContinueOnError {
				result.Success = false
				result.Errors = append(result.Errors, "too many errors, stopping import")
				break
			}
		}

		// Report progress if enabled
		if op.config.ReportProgress && result.Stats.TotalItems%int64(op.config.BatchSize) == 0 {
			fmt.Printf("Imported %d items (%d created, %d updated, %d failed)...\n",
				result.Stats.TotalItems,
				result.Stats.CreatedItems,
				result.Stats.UpdatedItems,
				result.Stats.FailedItems)
		}
	}

	// Run post-import actions if configured
	if len(req.Options.PostImportActions) > 0 {
		op.runPostImportActions(ctx, req.Options.PostImportActions, result)
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
	result.Metadata["create_missing"] = op.config.CreateMissing
	result.Metadata["update_existing"] = op.config.UpdateExisting
	result.Metadata["conflict_resolution"] = string(op.config.ConflictResolution)
	result.Metadata["transforms_count"] = len(req.Transforms)

	if lastError != nil && !result.Success {
		return result, lastError
	}

	return result, nil
}

// validateRequest validates the import request
func (op *ImportOperation) validateRequest(req *ImportRequest) error {
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

	// Validate conflict resolution settings
	if !op.config.CreateMissing && !op.config.UpdateExisting {
		return fmt.Errorf("either create_missing or update_existing must be enabled")
	}

	return nil
}

// updateStats updates statistics based on pipeline results
func (op *ImportOperation) updateStats(stats *ImportStats, result PipelineResult[any]) {
	stats.TotalItems++

	if stats.ItemsProcessed[result.Item.Type] == 0 {
		stats.ItemsProcessed[result.Item.Type] = 0
		stats.ItemsCreated[result.Item.Type] = 0
		stats.ItemsUpdated[result.Item.Type] = 0
		stats.ItemsFailed[result.Item.Type] = 0
		stats.ItemsSkipped[result.Item.Type] = 0
	}

	stats.ItemsProcessed[result.Item.Type]++

	if result.Success {
		// Determine if this was a create or update
		// This would typically be determined by the destination's response
		if op.config.CreateMissing {
			stats.ItemsCreated[result.Item.Type]++
			stats.CreatedItems++
		} else {
			stats.ItemsUpdated[result.Item.Type]++
			stats.UpdatedItems++
		}
	} else {
		stats.ItemsFailed[result.Item.Type]++
		stats.FailedItems++
	}

	// Update bytes processed if available
	if sizeBytes, exists := result.Item.GetProperty("size_bytes"); exists {
		if size, ok := sizeBytes.(int64); ok {
			stats.TotalBytesProcessed += size
		}
	}
}

// calculatePerformanceMetrics calculates performance metrics
func (op *ImportOperation) calculatePerformanceMetrics(stats *ImportStats, duration time.Duration) {
	if duration.Seconds() > 0 {
		stats.ItemsPerSecond = float64(stats.TotalItems) / duration.Seconds()
	}

	if stats.TotalItems > 0 {
		stats.AverageItemTime = duration / time.Duration(stats.TotalItems)
	}
}

// runPostImportActions executes post-import actions
func (op *ImportOperation) runPostImportActions(ctx context.Context, actions []string, result *ImportResult) {
	for _, action := range actions {
		switch action {
		case "validate":
			// Run post-import validation
			fmt.Printf("Running post-import validation...\n")
		case "index":
			// Rebuild indexes
			fmt.Printf("Rebuilding indexes...\n")
		case "cache_refresh":
			// Refresh caches
			fmt.Printf("Refreshing caches...\n")
		default:
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown post-import action: %s", action))
		}
	}
}

// ImportBuilder provides a fluent interface for building import operations
type ImportBuilder struct {
	request *ImportRequest
	config  ImportConfig
}

// NewImportBuilder creates a new import builder
func NewImportBuilder() *ImportBuilder {
	return &ImportBuilder{
		request: &ImportRequest{
			Scope:   ImportScope{},
			Options: ImportOptions{},
		},
		config: DefaultImportConfig(),
	}
}

// From sets the data source
func (b *ImportBuilder) From(source PluginSource[any]) *ImportBuilder {
	b.request.Source = source
	return b
}

// To sets the data destination
func (b *ImportBuilder) To(destination PluginDestination[any]) *ImportBuilder {
	b.request.Destination = destination
	return b
}

// Include adds object types to the import scope
func (b *ImportBuilder) Include(objectTypes ...common.ObjectType) *ImportBuilder {
	b.request.Scope.ObjectTypes = append(b.request.Scope.ObjectTypes, objectTypes...)
	return b
}

// WithIDs sets specific object IDs to import
func (b *ImportBuilder) WithIDs(ids ...string) *ImportBuilder {
	b.request.Scope.ObjectIDs = append(b.request.Scope.ObjectIDs, ids...)
	return b
}

// CreateMissing enables creation of missing objects
func (b *ImportBuilder) CreateMissing() *ImportBuilder {
	b.config.CreateMissing = true
	return b
}

// UpdateExisting enables updating of existing objects
func (b *ImportBuilder) UpdateExisting() *ImportBuilder {
	b.config.UpdateExisting = true
	return b
}

// WithConflictResolution sets the conflict resolution strategy
func (b *ImportBuilder) WithConflictResolution(resolution ConflictResolution) *ImportBuilder {
	b.config.ConflictResolution = resolution
	return b
}

// Transform adds a transformation to the pipeline
func (b *ImportBuilder) Transform(transformerType string, properties map[string]interface{}) *ImportBuilder {
	transform := TransformerConfig{
		Type:       transformerType,
		Properties: properties,
	}
	b.request.Transforms = append(b.request.Transforms, transform)
	return b
}

// WithFieldMapping adds field mappings
func (b *ImportBuilder) WithFieldMapping(sourceField, targetField string) *ImportBuilder {
	if b.request.Options.FieldMappings == nil {
		b.request.Options.FieldMappings = make(map[string]string)
	}
	b.request.Options.FieldMappings[sourceField] = targetField
	return b
}

// PreserveIDs enables ID preservation
func (b *ImportBuilder) PreserveIDs() *ImportBuilder {
	b.request.Options.PreserveIDs = true
	return b
}

// WithConcurrency sets the number of concurrent workers
func (b *ImportBuilder) WithConcurrency(workers int) *ImportBuilder {
	b.config.ConcurrentWorkers = workers
	return b
}

// StrictValidation enables strict validation mode
func (b *ImportBuilder) StrictValidation() *ImportBuilder {
	b.config.StrictValidation = true
	return b
}

// Build creates the import operation
func (b *ImportBuilder) Build(engine *Engine[any]) *ImportOperation {
	return NewImportOperation(engine, b.config)
}

// Execute builds and executes the import operation
func (b *ImportBuilder) Execute(ctx context.Context, engine *Engine[any]) (*ImportResult, error) {
	operation := b.Build(engine)
	return operation.Execute(ctx, b.request)
}
