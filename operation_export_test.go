package engine

// import (
// 	"context"
// 	"testing"
// 	"time"
// )

// func TestNewExportOperation(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination)

// 	if export == nil {
// 		t.Fatal("Expected export operation to be created")
// 	}
// }

// func TestExportOperation_WithTypes(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination).
// 		WithTypes(engine.ObjectTypePage, engine.ObjectTypeBlock)

// 	// Test that types were set (we can't directly access them, but can verify via execution)
// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	var count int
// 	for range results {
// 		count++
// 	}

// 	// Should have processed items (exact count depends on mock implementation)
// 	if count == 0 {
// 		t.Error("Expected some results from export")
// 	}
// }

// func TestExportOperation_WithDateRange(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	start := time.Now().Add(-24 * time.Hour)
// 	end := time.Now()

// 	export := operations.NewExportOperation(source, destination).
// 		WithDateRange(start, end)

// 	// Test execution
// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	for range results {
// 		// Just consume to complete the operation
// 	}
// }

// func TestExportOperation_WithFilter(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	filter := engine.Filter{
// 		Field:     "title",
// 		Operation: engine.FilterOpContains,
// 		Value:     "test",
// 	}

// 	export := operations.NewExportOperation(source, destination).
// 		WithFilter(filter)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	for range results {
// 		// Just consume to complete the operation
// 	}
// }

// func TestExportOperation_WithPagination(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	pagination := engine.PaginationOptions{
// 		PageSize: 10,
// 		MaxPages: 5,
// 	}

// 	export := operations.NewExportOperation(source, destination).
// 		WithPagination(pagination)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	for range results {
// 		// Just consume to complete the operation
// 	}
// }

// func TestExportOperation_WithTransformers(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	transformer1 := &MockTransformer{name: "transformer1"}
// 	transformer2 := &MockTransformer{name: "transformer2"}

// 	export := operations.NewExportOperation(source, destination).
// 		WithTransformers(transformer1, transformer2)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results and check transformations were applied
// 	var processedItems []engine.DataItem
// 	for result := range results {
// 		if result.Item != nil {
// 			processedItems = append(processedItems, *result.Item)
// 		}
// 	}

// 	// Check that transformations were applied
// 	for _, item := range processedItems {
// 		if transformer1Applied, exists := item.GetProperty("transformed_by_transformer1"); !exists || !transformer1Applied.(bool) {
// 			t.Error("Expected transformer1 to be applied")
// 		}
// 		if transformer2Applied, exists := item.GetProperty("transformed_by_transformer2"); !exists || !transformer2Applied.(bool) {
// 			t.Error("Expected transformer2 to be applied")
// 		}
// 	}
// }

// func TestExportOperation_WithValidators(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	validator := &MockValidator{name: "test-validator"}

// 	export := operations.NewExportOperation(source, destination).
// 		WithValidators(validator)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	for range results {
// 		// Just consume to complete the operation
// 	}
// }

// func TestExportOperation_WithOptions(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	options := engine.PipelineOptions{
// 		MaxWorkers:      3,
// 		BatchSize:       50,
// 		BufferSize:      500,
// 		Timeout:         15 * time.Minute,
// 		RateLimitRPS:    2.0,
// 		DryRun:          true,  // Should not actually write
// 		ContinueOnError: false, // Should stop on first error
// 	}

// 	export := operations.NewExportOperation(source, destination).
// 		WithOptions(options)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	var processedCount int
// 	for result := range results {
// 		processedCount++

// 		// In dry run mode, items should be processed but not written
// 		if result.Phase == engine.PhaseWrite && result.Status == engine.ProcessingStatusCompleted {
// 			// This shouldn't happen in dry run mode
// 			t.Error("Expected dry run to skip actual writing")
// 		}
// 	}

// 	if processedCount == 0 {
// 		t.Error("Expected some processing results even in dry run mode")
// 	}
// }

// func TestExportOperation_GetStatistics(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-2", map[string]interface{}{"title": "Page 2"}),
// 			*engine.NewDataItem(engine.ObjectTypeBlock, "block-1", map[string]interface{}{"content": "Block 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination)

// 	// Execute the export
// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume all results
// 	for range results {
// 		// Just consume to complete the operation
// 	}

// 	// Get statistics
// 	stats := export.GetStatistics()

// 	if stats.TotalItems == 0 {
// 		t.Error("Expected TotalItems to be greater than 0")
// 	}

// 	if stats.ProcessedItems == 0 {
// 		t.Error("Expected ProcessedItems to be greater than 0")
// 	}

// 	if stats.SuccessfulItems == 0 {
// 		t.Error("Expected SuccessfulItems to be greater than 0")
// 	}

// 	if stats.StartTime.IsZero() {
// 		t.Error("Expected StartTime to be set")
// 	}

// 	if stats.EndTime.IsZero() {
// 		t.Error("Expected EndTime to be set")
// 	}

// 	if stats.Duration == 0 {
// 		t.Error("Expected Duration to be greater than 0")
// 	}

// 	// Check type breakdown
// 	if stats.TypeBreakdown == nil {
// 		t.Error("Expected TypeBreakdown to be populated")
// 	}
// }

// func TestExportOperation_HierarchicalExport(t *testing.T) {
// 	// Create source with hierarchical data
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*func() *engine.DataItem {
// 				item := engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Parent Page"})
// 				item.AddRelationship(engine.RelationTypeChild, "block-1", engine.ObjectTypeBlock)
// 				item.AddRelationship(engine.RelationTypeChild, "block-2", engine.ObjectTypeBlock)
// 				return item
// 			}(),
// 			*func() *engine.DataItem {
// 				item := engine.NewDataItem(engine.ObjectTypeBlock, "block-1", map[string]interface{}{"content": "First block"})
// 				item.AddRelationship(engine.RelationTypeParent, "page-1", engine.ObjectTypePage)
// 				return item
// 			}(),
// 			*func() *engine.DataItem {
// 				item := engine.NewDataItem(engine.ObjectTypeBlock, "block-2", map[string]interface{}{"content": "Second block"})
// 				item.AddRelationship(engine.RelationTypeParent, "page-1", engine.ObjectTypePage)
// 				return item
// 			}(),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination).
// 		WithHierarchical(true).
// 		WithMaxDepth(2)

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	var itemsProcessed []engine.DataItem
// 	for result := range results {
// 		if result.Item != nil {
// 			itemsProcessed = append(itemsProcessed, *result.Item)
// 		}
// 	}

// 	// Should have processed all items with relationships preserved
// 	if len(itemsProcessed) == 0 {
// 		t.Error("Expected some items to be processed")
// 	}

// 	// Check that relationships are preserved
// 	for _, item := range itemsProcessed {
// 		if item.ID == "page-1" && len(item.Relationships) == 0 {
// 			t.Error("Expected parent page to have child relationships")
// 		}
// 	}
// }

// func TestExportOperation_ErrorHandling(t *testing.T) {
// 	source := &MockSource{
// 		shouldError:  true,
// 		errorMessage: "mock source error",
// 	}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination).
// 		WithOptions(engine.PipelineOptions{
// 			ContinueOnError: false, // Should stop on error
// 		})

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err == nil {
// 		t.Error("Expected error from failing source")
// 	}

// 	// Should not get any results due to immediate error
// 	if results != nil {
// 		for range results {
// 			// Shouldn't get here, but consume if we do
// 		}
// 	}
// }

// func TestExportOperation_ContinueOnError(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "good-page", map[string]interface{}{"title": "Good Page"}),
// 			*engine.NewDataItem(engine.ObjectTypePage, "bad-page", "invalid-data"), // This should cause validation error
// 			*engine.NewDataItem(engine.ObjectTypePage, "another-good-page", map[string]interface{}{"title": "Another Good Page"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination).
// 		WithOptions(engine.PipelineOptions{
// 			ContinueOnError: true, // Should continue despite errors
// 		})

// 	ctx := context.Background()
// 	results, err := export.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error with ContinueOnError=true, got %v", err)
// 	}

// 	var errorCount, successCount int
// 	for result := range results {
// 		if result.Error != nil {
// 			errorCount++
// 		} else if result.Status == engine.ProcessingStatusCompleted {
// 			successCount++
// 		}
// 	}

// 	if errorCount == 0 {
// 		t.Error("Expected some errors due to invalid data")
// 	}

// 	if successCount == 0 {
// 		t.Error("Expected some successful items")
// 	}

// 	// Check statistics include error count
// 	stats := export.GetStatistics()
// 	if stats.FailedItems == 0 {
// 		t.Error("Expected FailedItems to be greater than 0")
// 	}
// }

// func TestExportOperation_CancelContext(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-2", map[string]interface{}{"title": "Page 2"}),
// 		},
// 		delay: 100 * time.Millisecond, // Add delay to allow cancellation
// 	}
// 	destination := &MockDestination{}

// 	export := operations.NewExportOperation(source, destination)

// 	// Create context with short timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
// 	defer cancel()

// 	results, err := export.Execute(ctx)
// 	if err == nil {
// 		t.Error("Expected context timeout error")
// 	}

// 	if results != nil {
// 		// Consume any results that came through before timeout
// 		for range results {
// 		}
// 	}
// }

// // Mock implementations for testing

// type MockSource struct {
// 	items        []engine.DataItem
// 	shouldError  bool
// 	errorMessage string
// 	delay        time.Duration
// }

// func (s *MockSource) Read(ctx context.Context, req *engine.ReadRequest) (<-chan engine.DataItem, error) {
// 	if s.shouldError {
// 		return nil, stringError(s.errorMessage)
// 	}

// 	results := make(chan engine.DataItem, len(s.items))

// 	go func() {
// 		defer close(results)
// 		for _, item := range s.items {
// 			if s.delay > 0 {
// 				select {
// 				case <-time.After(s.delay):
// 				case <-ctx.Done():
// 					return
// 				}
// 			}

// 			select {
// 			case results <- item:
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()

// 	return results, nil
// }

// func (s *MockSource) Validate(req *engine.ReadRequest) error {
// 	if s.shouldError {
// 		return stringError(s.errorMessage)
// 	}
// 	return nil
// }

// func (s *MockSource) SupportsType(objType engine.ObjectType) bool {
// 	return true
// }

// func (s *MockSource) Config() engine.SourceConfig {
// 	return engine.SourceConfig{
// 		Type: "mock",
// 		Name: "Mock Source",
// 	}
// }

// func (s *MockSource) Close() error {
// 	return nil
// }

// type MockDestination struct {
// 	items []engine.DataItem
// }

// func (d *MockDestination) Write(ctx context.Context, item engine.DataItem) error {
// 	d.items = append(d.items, item)
// 	return nil
// }

// func (d *MockDestination) Batch(ctx context.Context, items []engine.DataItem) error {
// 	d.items = append(d.items, items...)
// 	return nil
// }

// func (d *MockDestination) SupportsType(objType engine.ObjectType) bool {
// 	return true
// }

// func (d *MockDestination) Validate(item engine.DataItem) error {
// 	// Simulate validation error for items with string data (should be map)
// 	if _, ok := item.Data.(string); ok && item.Data.(string) == "invalid-data" {
// 		return stringError("invalid data type")
// 	}
// 	return nil
// }

// func (d *MockDestination) Config() engine.DestinationConfig {
// 	return engine.DestinationConfig{
// 		Type: "mock",
// 		Name: "Mock Destination",
// 	}
// }

// func (d *MockDestination) Close() error {
// 	return nil
// }

// type MockTransformer struct {
// 	name string
// }

// func (t *MockTransformer) Transform(ctx context.Context, item engine.DataItem) (engine.DataItem, error) {
// 	// Mark item as transformed by this transformer
// 	item.SetProperty("transformed_by_"+t.name, true)
// 	return item, nil
// }

// func (t *MockTransformer) GetName() string {
// 	return t.name
// }

// func (t *MockTransformer) GetDescription() string {
// 	return "Mock transformer for testing"
// }

// type MockValidator struct {
// 	name string
// }

// func (v *MockValidator) Validate(ctx context.Context, item engine.DataItem) engine.ValidationState {
// 	return engine.ValidationState{
// 		IsValid:     true,
// 		Errors:      []engine.ValidationError{},
// 		Warnings:    []engine.ValidationWarning{},
// 		ValidatedAt: time.Now(),
// 		Rules:       []string{v.name},
// 	}
// }

// func (v *MockValidator) GetName() string {
// 	return v.name
// }

// func (v *MockValidator) GetDescription() string {
// 	return "Mock validator for testing"
// }

// // Helper to convert string to error (Go doesn't have a direct conversion)
// type stringError string

// func (e stringError) Error() string { return string(e) }
