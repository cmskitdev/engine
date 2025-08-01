package engine

// import (
// 	"context"
// 	"testing"
// )

// func TestNewImportOperation(t *testing.T) {
// 	source := &MockSource{}
// 	destination := &MockDestination{}

// 	importOp := NewImportOperation(source, destination)

// 	if importOp == nil {
// 		t.Fatal("Expected import operation to be created")
// 	}
// }

// func TestImportOperation_WithConflictResolution(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithConflictResolution(operations.ConflictResolutionOverwrite)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	for range results {
// 		// Just consume to complete the operation
// 	}

// 	stats := importOp.GetStatistics()
// 	if stats.TotalItems == 0 {
// 		t.Error("Expected some items to be processed")
// 	}
// }

// func TestImportOperation_WithFieldMapping(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{
// 				"old_title":   "Page 1",
// 				"old_content": "Content 1",
// 			}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	fieldMappings := map[string]string{
// 		"old_title":   "title",
// 		"old_content": "content",
// 	}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithFieldMapping(fieldMappings)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results and check field mappings were applied
// 	var processedItems []engine.DataItem
// 	for result := range results {
// 		if result.Item != nil {
// 			processedItems = append(processedItems, *result.Item)
// 		}
// 	}

// 	// Check that field mappings were applied
// 	for _, item := range processedItems {
// 		if data, ok := item.Data.(map[string]interface{}); ok {
// 			if data["title"] != "Page 1" {
// 				t.Error("Expected old_title to be mapped to title")
// 			}
// 			if data["content"] != "Content 1" {
// 				t.Error("Expected old_content to be mapped to content")
// 			}
// 			// Old fields should be removed
// 			if data["old_title"] != nil {
// 				t.Error("Expected old_title to be removed after mapping")
// 			}
// 			if data["old_content"] != nil {
// 				t.Error("Expected old_content to be removed after mapping")
// 			}
// 		}
// 	}
// }

// func TestImportOperation_WithDependencyResolution(t *testing.T) {
// 	// Create items with dependencies
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*func() *engine.DataItem {
// 				// Block depends on page
// 				item := engine.NewDataItem(engine.ObjectTypeBlock, "block-1", map[string]interface{}{"content": "Block 1"})
// 				item.AddDependency("page-1") // Depends on page-1
// 				return item
// 			}(),
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithDependencyResolution(true)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	var processOrder []string
// 	for result := range results {
// 		if result.Item != nil && result.Phase == engine.PhaseWrite && result.Status == engine.ProcessingStatusCompleted {
// 			processOrder = append(processOrder, result.Item.ID)
// 		}
// 	}

// 	// Page should be processed before block due to dependency
// 	if len(processOrder) >= 2 {
// 		pageIndex := -1
// 		blockIndex := -1
// 		for i, id := range processOrder {
// 			if id == "page-1" {
// 				pageIndex = i
// 			}
// 			if id == "block-1" {
// 				blockIndex = i
// 			}
// 		}
// 		if pageIndex == -1 || blockIndex == -1 {
// 			t.Error("Expected both page and block to be processed")
// 		} else if pageIndex > blockIndex {
// 			t.Error("Expected page to be processed before block due to dependency")
// 		}
// 	}
// }

// func TestImportOperation_WithValidationBeforeImport(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "valid-page", map[string]interface{}{"title": "Valid Page"}),
// 			*engine.NewDataItem(engine.ObjectTypePage, "invalid-page", "invalid-data"), // Should fail validation
// 		},
// 	}
// 	destination := &MockDestination{}

// 	validator := &MockValidator{name: "strict-validator"}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithValidationBeforeImport(true).
// 		WithValidators(validator)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	var validItems, invalidItems int
// 	for result := range results {
// 		if result.Error != nil {
// 			invalidItems++
// 		} else if result.Status == engine.ProcessingStatusCompleted {
// 			validItems++
// 		}
// 	}

// 	if validItems == 0 {
// 		t.Error("Expected some valid items to be processed")
// 	}

// 	// Note: The actual validation behavior depends on the validator implementation
// 	// Our mock validator currently always returns valid, so we don't expect validation failures
// }

// func TestImportOperation_WithBatchSize(t *testing.T) {
// 	items := make([]engine.DataItem, 10)
// 	for i := 0; i < 10; i++ {
// 		items[i] = *engine.NewDataItem(engine.ObjectTypePage,
// 			"page-"+string(rune('0'+i)),
// 			map[string]interface{}{"title": "Page " + string(rune('0'+i))})
// 	}

// 	source := &MockSource{items: items}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithBatchSize(3) // Process in batches of 3

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results
// 	var batchCount int
// 	for result := range results {
// 		if result.Phase == engine.PhaseWrite && result.Status == engine.ProcessingStatusCompleted {
// 			batchCount++
// 		}
// 	}

// 	if batchCount == 0 {
// 		t.Error("Expected some batches to be processed")
// 	}
// }

// func TestImportOperation_WithIDMapping(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "external-page-1", map[string]interface{}{"title": "External Page 1"}),
// 			*engine.NewDataItem(engine.ObjectTypePage, "external-page-2", map[string]interface{}{"title": "External Page 2"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	idMappings := map[string]string{
// 		"external-page-1": "internal-page-1",
// 		"external-page-2": "internal-page-2",
// 	}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithIDMapping(idMappings)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results and check ID mappings
// 	var processedItems []engine.DataItem
// 	for result := range results {
// 		if result.Item != nil {
// 			processedItems = append(processedItems, *result.Item)
// 		}
// 	}

// 	// Check that IDs were mapped
// 	foundMappedIDs := make(map[string]bool)
// 	for _, item := range processedItems {
// 		if item.ID == "internal-page-1" || item.ID == "internal-page-2" {
// 			foundMappedIDs[item.ID] = true
// 		}
// 	}

// 	if !foundMappedIDs["internal-page-1"] {
// 		t.Error("Expected external-page-1 to be mapped to internal-page-1")
// 	}
// 	if !foundMappedIDs["internal-page-2"] {
// 		t.Error("Expected external-page-2 to be mapped to internal-page-2")
// 	}
// }

// func TestImportOperation_WithTransformers(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	transformer := &MockTransformer{name: "import-transformer"}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithTransformers(transformer)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume results and verify transformation
// 	var transformedItems []engine.DataItem
// 	for result := range results {
// 		if result.Item != nil {
// 			transformedItems = append(transformedItems, *result.Item)
// 		}
// 	}

// 	// Check that transformation was applied
// 	for _, item := range transformedItems {
// 		if transformed, exists := item.GetProperty("transformed_by_import-transformer"); !exists || !transformed.(bool) {
// 			t.Error("Expected transformer to be applied")
// 		}
// 	}
// }

// func TestImportOperation_ConflictHandling_Skip(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "New Title"}),
// 		},
// 	}

// 	// Mock destination that simulates existing items
// 	destination := &MockDestinationWithConflicts{
// 		existingItems: map[string]engine.DataItem{
// 			"page-1": *engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Existing Title"}),
// 		},
// 	}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithConflictResolution(operations.ConflictResolutionSkip)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	var skippedCount, processedCount int
// 	for result := range results {
// 		if result.Status == engine.ProcessingStatusSkipped {
// 			skippedCount++
// 		} else if result.Status == engine.ProcessingStatusCompleted {
// 			processedCount++
// 		}
// 	}

// 	if skippedCount == 0 {
// 		t.Error("Expected some items to be skipped due to conflicts")
// 	}

// 	// Check statistics
// 	stats := importOp.GetStatistics()
// 	if stats.ConflictsResolved == 0 {
// 		t.Error("Expected some conflicts to be recorded")
// 	}
// }

// func TestImportOperation_ConflictHandling_Overwrite(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "New Title"}),
// 		},
// 	}

// 	destination := &MockDestinationWithConflicts{
// 		existingItems: map[string]engine.DataItem{
// 			"page-1": *engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Existing Title"}),
// 		},
// 	}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithConflictResolution(operations.ConflictResolutionOverwrite)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	var overwrittenCount int
// 	for result := range results {
// 		if result.Status == engine.ProcessingStatusCompleted {
// 			overwrittenCount++
// 		}
// 	}

// 	if overwrittenCount == 0 {
// 		t.Error("Expected items to be overwritten")
// 	}
// }

// func TestImportOperation_GetStatistics(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 			*engine.NewDataItem(engine.ObjectTypeBlock, "block-1", map[string]interface{}{"content": "Block 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination)

// 	// Execute the import
// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	// Consume all results
// 	for range results {
// 		// Just consume to complete the operation
// 	}

// 	// Get statistics
// 	stats := importOp.GetStatistics()

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

// 	// Check import-specific statistics
// 	if stats.TypeBreakdown == nil {
// 		t.Error("Expected TypeBreakdown to be populated")
// 	}

// 	if stats.FieldMappings == nil {
// 		stats.FieldMappings = make(map[string]string) // Initialize if nil
// 	}

// 	if stats.ConflictsResolved < 0 {
// 		t.Error("Expected ConflictsResolved to be non-negative")
// 	}
// }

// func TestImportOperation_ErrorHandling(t *testing.T) {
// 	source := &MockSource{
// 		shouldError:  true,
// 		errorMessage: "mock import source error",
// 	}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err == nil {
// 		t.Error("Expected error from failing source")
// 	}

// 	// Should not get results due to error
// 	if results != nil {
// 		for range results {
// 			// Shouldn't get here, but consume if we do
// 		}
// 	}
// }

// func TestImportOperation_DryRun(t *testing.T) {
// 	source := &MockSource{
// 		items: []engine.DataItem{
// 			*engine.NewDataItem(engine.ObjectTypePage, "page-1", map[string]interface{}{"title": "Page 1"}),
// 		},
// 	}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination).
// 		WithOptions(engine.PipelineOptions{
// 			DryRun: true,
// 		})

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	var dryRunCount int
// 	for result := range results {
// 		if result.Phase == engine.PhaseWrite {
// 			// In dry run, writes should be simulated but not executed
// 			dryRunCount++
// 		}
// 	}

// 	if dryRunCount == 0 {
// 		t.Error("Expected some dry run operations")
// 	}

// 	// Destination should not have received any actual writes in dry run mode
// 	if len(destination.items) > 0 {
// 		t.Error("Expected no items to be written in dry run mode")
// 	}
// }

// // Mock destination that simulates conflicts
// type MockDestinationWithConflicts struct {
// 	MockDestination
// 	existingItems map[string]engine.DataItem
// }

// func (d *MockDestinationWithConflicts) Write(ctx context.Context, item engine.DataItem) error {
// 	// Check for conflicts
// 	if existing, exists := d.existingItems[item.ID]; exists {
// 		// Simulate conflict detection
// 		if existing.Data != item.Data {
// 			// This would normally trigger conflict resolution
// 			// For testing, we'll just proceed based on the resolution strategy
// 		}
// 	}

// 	return d.MockDestination.Write(ctx, item)
// }

// func (d *MockDestinationWithConflicts) HasItem(id string) bool {
// 	_, exists := d.existingItems[id]
// 	return exists
// }

// func TestConflictResolutionConstants(t *testing.T) {
// 	resolutions := []operations.ConflictResolution{
// 		operations.ConflictResolutionSkip,
// 		operations.ConflictResolutionOverwrite,
// 		operations.ConflictResolutionMerge,
// 		operations.ConflictResolutionRename,
// 		operations.ConflictResolutionError,
// 	}

// 	expectedValues := []string{"skip", "overwrite", "merge", "rename", "error"}

// 	for i, resolution := range resolutions {
// 		if string(resolution) != expectedValues[i] {
// 			t.Errorf("Expected resolution %d to be %s, got %s", i, expectedValues[i], string(resolution))
// 		}
// 	}
// }

// func TestImportOperation_ProgressTracking(t *testing.T) {
// 	items := make([]engine.DataItem, 5)
// 	for i := 0; i < 5; i++ {
// 		items[i] = *engine.NewDataItem(engine.ObjectTypePage,
// 			"page-"+string(rune('0'+i)),
// 			map[string]interface{}{"title": "Page " + string(rune('0'+i))})
// 	}

// 	source := &MockSource{items: items}
// 	destination := &MockDestination{}

// 	importOp := operations.NewImportOperation(source, destination)

// 	ctx := context.Background()
// 	results, err := importOp.Execute(ctx)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	var progressUpdates []engine.PipelineResult
// 	for result := range results {
// 		progressUpdates = append(progressUpdates, result)
// 	}

// 	if len(progressUpdates) == 0 {
// 		t.Error("Expected progress updates")
// 	}

// 	// Should have updates for different phases
// 	phasesSeen := make(map[engine.ProcessingPhase]bool)
// 	for _, update := range progressUpdates {
// 		phasesSeen[update.Phase] = true
// 	}

// 	expectedPhases := []engine.ProcessingPhase{
// 		engine.PhaseRead,
// 		engine.PhaseValidate,
// 		engine.PhaseTransform,
// 		engine.PhaseWrite,
// 	}

// 	for _, phase := range expectedPhases {
// 		if !phasesSeen[phase] {
// 			t.Errorf("Expected to see progress updates for phase %s", phase)
// 		}
// 	}
// }
