package engine

import (
	"context"
	"time"

	"github.com/cmskitdev/common"
)

// Pipeline is the main interface for processing data in either direction (export/import)
type Pipeline[T any] interface {
	// Process data through the pipeline
	Process(ctx context.Context, req *PipelineRequest[T]) (<-chan PipelineResult[T], error)

	// Get pipeline configuration
	Config() PipelineConfig

	// Close the pipeline and clean up resources
	Close() error
}

// Direction specifies the data flow direction
type Direction string

const (
	DirectionExport Direction = "export"
	DirectionImport Direction = "import"
	DirectionSync   Direction = "sync"
)

// PipelineRequest defines a complete data processing request
type PipelineRequest[T any] struct {
	// Processing direction
	Direction Direction `json:"direction"`

	// Data source configuration
	Source PluginSource[T] `json:"source"`

	// Data destination configuration
	Destination PluginDestination[T] `json:"destination"`

	// Transformation configuration
	Transform TransformConfig `json:"transform"`

	// Processing options
	Options PipelineOptions `json:"options"`

	// Validation configuration
	Validation ValidationConfig `json:"validation"`

	// Request metadata
	Metadata RequestMetadata `json:"metadata"`
}

// PipelineResult represents the result of processing a single data item
type PipelineResult[T any] struct {
	// The processed data item
	Item DataItemContainer[T] `json:"item"`

	// Processing phase where this result was generated
	Phase ProcessingPhase `json:"phase"`

	// Success indicator
	Success bool `json:"success"`

	// Error if processing failed
	Error error `json:"error,omitempty"`

	// Processing metadata
	Metadata ResultMetadata `json:"metadata"`
}

// PipelineOptions controls how the pipeline processes data
type PipelineOptions struct {
	// Concurrency settings
	MaxWorkers int           `json:"max_workers"`
	BatchSize  int           `json:"batch_size"`
	BufferSize int           `json:"buffer_size"`
	Timeout    time.Duration `json:"timeout"`

	// Error handling
	ContinueOnError bool          `json:"continue_on_error"`
	MaxErrors       int           `json:"max_errors"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryBackoff    time.Duration `json:"retry_backoff"`

	// Progress tracking
	EnableProgress   bool          `json:"enable_progress"`
	ProgressInterval time.Duration `json:"progress_interval"`

	// Resource management
	RateLimitRPS  float64 `json:"rate_limit_rps"`
	MemoryLimitMB int     `json:"memory_limit_mb"`

	// Processing flags
	DryRun            bool `json:"dry_run"`
	SkipValidation    bool `json:"skip_validation"`
	EnableCompression bool `json:"enable_compression"`
}

// DefaultPipelineOptions returns sensible defaults
func DefaultPipelineOptions() PipelineOptions {
	return PipelineOptions{
		MaxWorkers:        5,
		BatchSize:         100,
		BufferSize:        1000,
		Timeout:           30 * time.Minute,
		ContinueOnError:   true,
		MaxErrors:         10,
		RetryAttempts:     3,
		RetryBackoff:      time.Second,
		EnableProgress:    true,
		ProgressInterval:  10 * time.Second,
		RateLimitRPS:      3.0, // Notion API limit
		MemoryLimitMB:     512,
		DryRun:            false,
		SkipValidation:    false,
		EnableCompression: false,
	}
}

// RequestMetadata contains metadata about the pipeline request
type RequestMetadata struct {
	RequestID string            `json:"request_id"`
	UserID    string            `json:"user_id,omitempty"`
	Source    string            `json:"source,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// ResultMetadata contains metadata about processing results
type ResultMetadata struct {
	Time       time.Time              `json:"time"`
	Duration   time.Duration          `json:"duration"`
	Attempts   int                    `json:"attempts"`
	WorkerID   string                 `json:"worker_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ProcessingPhase represents the current phase of data processing
type ProcessingPhase string

const (
	PhaseRead      ProcessingPhase = "read"
	PhaseTransform ProcessingPhase = "transform"
	PhaseValidate  ProcessingPhase = "validate"
	PhaseRoute     ProcessingPhase = "route"
	PhaseWrite     ProcessingPhase = "write"
	PhaseComplete  ProcessingPhase = "complete"
	PhaseError     ProcessingPhase = "error"
)

// PipelineConfig represents the configuration of a pipeline
type PipelineConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// PluginSource interface for reading data from various sources
type PluginSource[T any] interface {
	// Read data items from the source
	Read(ctx context.Context, req *ReadRequest) (<-chan DataItemContainer[T], error)

	// Validate that the read request is valid for this source
	Validate(req *ReadRequest) error

	// Check if this source supports the given object type
	SupportsType(objType common.ObjectType) bool

	// Get source configuration
	Config() SourceConfig

	// Close the source and clean up resources
	Close() error
}

// PluginDestination interface for writing data to various destinations
type PluginDestination[T any] interface {
	// Write a single data item to the destination
	Write(ctx context.Context, item DataItemContainer[T]) error

	// Write multiple data items in a batch (optional optimization)
	Batch(ctx context.Context, items []DataItemContainer[T]) error

	// Check if this destination supports the given object type
	SupportsType(objType common.ObjectType) bool

	// Validate that the data item is valid for this destination
	Validate(item DataItemContainer[T]) error

	// Get destination configuration
	Config() DestinationConfig

	// Close the destination and clean up resources
	Close() error
}

// ReadRequest specifies what data to read from a source
type ReadRequest struct {
	// Object types to read
	Types []common.ObjectType `json:"types,omitempty"`

	// Specific object IDs to read
	IDs []string `json:"ids,omitempty"`

	// Filters to apply
	Filters []Filter `json:"filters,omitempty"`

	// Pagination options
	Pagination *PaginationOptions `json:"pagination,omitempty"`

	// Additional options specific to the source
	Options map[string]interface{} `json:"options,omitempty"`
}

// PaginationOptions for reading data in chunks
type PaginationOptions struct {
	PageSize    int    `json:"page_size"`
	StartCursor string `json:"start_cursor,omitempty"`
	MaxPages    int    `json:"max_pages,omitempty"`
}

// Filter represents a data filter
type Filter struct {
	Field     string      `json:"field"`
	Operation FilterOp    `json:"operation"`
	Value     interface{} `json:"value"`
}

// FilterOp represents filter operations
type FilterOp string

const (
	FilterOpEquals      FilterOp = "equals"
	FilterOpNotEquals   FilterOp = "not_equals"
	FilterOpContains    FilterOp = "contains"
	FilterOpNotContains FilterOp = "not_contains"
	FilterOpStartsWith  FilterOp = "starts_with"
	FilterOpEndsWith    FilterOp = "ends_with"
	FilterOpGreaterThan FilterOp = "greater_than"
	FilterOpLessThan    FilterOp = "less_than"
	FilterOpIsEmpty     FilterOp = "is_empty"
	FilterOpIsNotEmpty  FilterOp = "is_not_empty"
)

// SourceConfig represents configuration for a data source
type SourceConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

// DestinationConfig represents configuration for a data destination
type DestinationConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

// TransformConfig specifies how data should be transformed
type TransformConfig struct {
	// List of transformers to apply in order
	Transformers []TransformerConfig `json:"transformers"`

	// Global transformation options
	Options TransformOptions `json:"options"`
}

// TransformerConfig configures a single transformer
type TransformerConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TransformOptions controls transformation behavior
type TransformOptions struct {
	SkipOnError      bool `json:"skip_on_error"`
	StrictMode       bool `json:"strict_mode"`
	PreserveMetadata bool `json:"preserve_metadata"`
}

// ValidationConfig specifies validation rules
type ValidationConfig struct {
	// List of validators to apply
	Validators []ValidatorConfig `json:"validators"`

	// Validation options
	Options ValidationOptions `json:"options"`
}

// ValidatorConfig configures a single validator
type ValidatorConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ValidationOptions controls validation behavior
type ValidationOptions struct {
	StopOnFirstError bool `json:"stop_on_first_error"`
	StrictMode       bool `json:"strict_mode"`
	SkipWarnings     bool `json:"skip_warnings"`
}
