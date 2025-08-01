package engine

import (
	"time"
)

// ObjectType represents the type of data object
type ObjectType string

const (
	ObjectTypeDatabase  ObjectType = "database"
	ObjectTypePage      ObjectType = "page"
	ObjectTypeBlock     ObjectType = "block"
	ObjectTypeComment   ObjectType = "comment"
	ObjectTypeUser      ObjectType = "user"
	ObjectTypeFile      ObjectType = "file"
	ObjectTypeWorkspace ObjectType = "workspace"
	ObjectTypeProperty  ObjectType = "property"
	ObjectTypeRelation  ObjectType = "relation"
	ObjectTypeGeneric   ObjectType = "generic"
)

// DataItem represents a universal data item that can flow through the pipeline
type DataItemContainer struct {
	// Core identification
	Type ObjectType `json:"type"`
	ID   string     `json:"id"`

	// The actual data payload
	Data interface{} `json:"data"`

	// Item metadata
	Metadata ItemMetadata `json:"metadata"`

	// Relationships to other items
	Relationships []Relationship `json:"relationships,omitempty"`

	// Dependencies that must be processed before this item
	Dependencies []string `json:"dependencies,omitempty"`

	// Tags for categorization and filtering
	Tags []string `json:"tags,omitempty"`
}

// ItemMetadata contains metadata about a data item
type ItemMetadata struct {
	// Source information
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id,omitempty"`
	OriginalID string `json:"original_id,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	ModifiedAt  time.Time  `json:"modified_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// Processing state
	ValidationState ValidationState `json:"validation_state"`
	TransformState  TransformState  `json:"transform_state"`
	ProcessingState ProcessingState `json:"processing_state"`

	// Additional properties
	Properties map[string]interface{} `json:"properties,omitempty"`

	// Version information
	Version  string `json:"version,omitempty"`
	Checksum string `json:"checksum,omitempty"`

	// Size information
	SizeBytes int64 `json:"size_bytes,omitempty"`
	ItemCount int   `json:"item_count,omitempty"` // For containers like pages with blocks
}

// ValidationState represents the validation status of an item
type ValidationState struct {
	IsValid     bool                `json:"is_valid"`
	Errors      []ValidationError   `json:"errors,omitempty"`
	Warnings    []ValidationWarning `json:"warnings,omitempty"`
	ValidatedAt time.Time           `json:"validated_at"`
	Rules       []string            `json:"rules,omitempty"` // Names of rules applied
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Field    string                 `json:"field,omitempty"`
	Severity ValidationSeverity     `json:"severity"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Field   string                 `json:"field,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ValidationSeverity represents the severity of a validation issue
type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "error"
	ValidationSeverityWarning ValidationSeverity = "warning"
	ValidationSeverityInfo    ValidationSeverity = "info"
)

// TransformState represents the transformation status of an item
type TransformState struct {
	IsTransformed   bool            `json:"is_transformed"`
	TransformedAt   time.Time       `json:"transformed_at"`
	TransformerUsed string          `json:"transformer_used,omitempty"`
	OriginalFormat  string          `json:"original_format,omitempty"`
	TargetFormat    string          `json:"target_format,omitempty"`
	TransformLog    []TransformStep `json:"transform_log,omitempty"`
}

// TransformStep represents a single transformation step
type TransformStep struct {
	StepName   string                 `json:"step_name"`
	Applied    bool                   `json:"applied"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ProcessingState represents the overall processing status
type ProcessingState struct {
	Phase       ProcessingPhase  `json:"phase"`
	Status      ProcessingStatus `json:"status"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Attempts    int              `json:"attempts"`
	LastError   string           `json:"last_error,omitempty"`
}

// ProcessingStatus represents the status of processing
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusInProgress ProcessingStatus = "in_progress"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
	ProcessingStatusSkipped    ProcessingStatus = "skipped"
	ProcessingStatusRetrying   ProcessingStatus = "retrying"
)

// Relationship represents a relationship between data items
type Relationship struct {
	Type          RelationType           `json:"type"`
	TargetID      string                 `json:"target_id"`
	TargetType    ObjectType             `json:"target_type,omitempty"`
	Context       string                 `json:"context,omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
	Bidirectional bool                   `json:"bidirectional,omitempty"`
}

// RelationType represents different types of relationships between items
type RelationType string

const (
	RelationTypeParent     RelationType = "parent"
	RelationTypeChild      RelationType = "child"
	RelationTypeReference  RelationType = "reference"
	RelationTypeDependency RelationType = "dependency"
	RelationTypeContains   RelationType = "contains"
	RelationTypeBelongsTo  RelationType = "belongs_to"
	RelationTypeRelated    RelationType = "related"
	RelationTypeLinked     RelationType = "linked"
	RelationTypeMentions   RelationType = "mentions"
	RelationTypeAnnotates  RelationType = "annotates"
)

// DataCollection represents a collection of related data items
type DataCollection struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Items       []DataItemContainer `json:"items"`
	Metadata    CollectionMetadata  `json:"metadata"`
	Schema      *CollectionSchema   `json:"schema,omitempty"`
}

// CollectionMetadata contains metadata about a collection
type CollectionMetadata struct {
	CreatedAt  time.Time              `json:"created_at"`
	ModifiedAt time.Time              `json:"modified_at"`
	ItemCount  int                    `json:"item_count"`
	TotalSize  int64                  `json:"total_size"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
}

// CollectionSchema defines the expected structure of items in a collection
type CollectionSchema struct {
	Version     string                 `json:"version"`
	ObjectTypes []ObjectType           `json:"object_types"`
	Fields      map[string]FieldSchema `json:"fields"`
	Required    []string               `json:"required,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// FieldSchema defines the schema for a specific field
type FieldSchema struct {
	Type        string            `json:"type"`
	Description string            `json:"description,omitempty"`
	Required    bool              `json:"required"`
	Default     interface{}       `json:"default,omitempty"`
	Constraints []FieldConstraint `json:"constraints,omitempty"`
}

// FieldConstraint represents a constraint on a field value
type FieldConstraint struct {
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Message string      `json:"message,omitempty"`
}

// Common constraint types
const (
	ConstraintTypeMinLength = "min_length"
	ConstraintTypeMaxLength = "max_length"
	ConstraintTypePattern   = "pattern"
	ConstraintTypeMinValue  = "min_value"
	ConstraintTypeMaxValue  = "max_value"
	ConstraintTypeOneOf     = "one_of"
	ConstraintTypeNotEmpty  = "not_empty"
	ConstraintTypeUnique    = "unique"
)

// Helper methods for DataItem

// IsValid checks if the item passed validation
func (item *DataItemContainer) IsValid() bool {
	return item.Metadata.ValidationState.IsValid
}

// HasErrors checks if the item has validation errors
func (item *DataItemContainer) HasErrors() bool {
	return len(item.Metadata.ValidationState.Errors) > 0
}

// HasWarnings checks if the item has validation warnings
func (item *DataItemContainer) HasWarnings() bool {
	return len(item.Metadata.ValidationState.Warnings) > 0
}

// IsTransformed checks if the item has been transformed
func (item *DataItemContainer) IsTransformed() bool {
	return item.Metadata.TransformState.IsTransformed
}

// IsProcessed checks if the item processing is complete
func (item *DataItemContainer) IsProcessed() bool {
	return item.Metadata.ProcessingState.Status == ProcessingStatusCompleted
}

// HasDependencies checks if the item has dependencies
func (item *DataItemContainer) HasDependencies() bool {
	return len(item.Dependencies) > 0
}

// AddRelationship adds a relationship to the item
func (item *DataItemContainer) AddRelationship(relType RelationType, targetID string, targetType ObjectType) {
	if item.Relationships == nil {
		item.Relationships = make([]Relationship, 0)
	}

	item.Relationships = append(item.Relationships, Relationship{
		Type:       relType,
		TargetID:   targetID,
		TargetType: targetType,
	})
}

// AddDependency adds a dependency to the item
func (item *DataItemContainer) AddDependency(dependencyID string) {
	if item.Dependencies == nil {
		item.Dependencies = make([]string, 0)
	}

	// Check if dependency already exists
	for _, dep := range item.Dependencies {
		if dep == dependencyID {
			return
		}
	}

	item.Dependencies = append(item.Dependencies, dependencyID)
}

// SetProperty sets a metadata property
func (item *DataItemContainer) SetProperty(key string, value interface{}) {
	if item.Metadata.Properties == nil {
		item.Metadata.Properties = make(map[string]interface{})
	}
	item.Metadata.Properties[key] = value
}

// GetProperty gets a metadata property
func (item *DataItemContainer) GetProperty(key string) (interface{}, bool) {
	if item.Metadata.Properties == nil {
		return nil, false
	}
	value, exists := item.Metadata.Properties[key]
	return value, exists
}

// AddTag adds a tag to the item
func (item *DataItemContainer) AddTag(tag string) {
	if item.Tags == nil {
		item.Tags = make([]string, 0)
	}

	// Check if tag already exists
	for _, t := range item.Tags {
		if t == tag {
			return
		}
	}

	item.Tags = append(item.Tags, tag)
}

// HasTag checks if the item has a specific tag
func (item *DataItemContainer) HasTag(tag string) bool {
	for _, t := range item.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// NewDataItem creates a new data item with basic metadata
func NewDataItem(objType ObjectType, id string, data interface{}) *DataItemContainer {
	now := time.Now()

	return &DataItemContainer{
		Type: objType,
		ID:   id,
		Data: data,
		Metadata: ItemMetadata{
			CreatedAt:  now,
			ModifiedAt: now,
			ValidationState: ValidationState{
				IsValid: false, // Must be validated explicitly
			},
			TransformState: TransformState{
				IsTransformed: false,
			},
			ProcessingState: ProcessingState{
				Phase:     PhaseRead,
				Status:    ProcessingStatusPending,
				StartedAt: now,
				Attempts:  0,
			},
		},
		Relationships: make([]Relationship, 0),
		Dependencies:  make([]string, 0),
		Tags:          make([]string, 0),
	}
}
