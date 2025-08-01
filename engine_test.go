package engine

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()

	if config.Name != "default-engine" {
		t.Errorf("Expected Name to be default-engine, got %s", config.Name)
	}

	if config.MaxConcurrentItems != 10 {
		t.Errorf("Expected MaxConcurrentItems to be 10, got %d", config.MaxConcurrentItems)
	}

	if config.DefaultTimeout != 30*time.Minute {
		t.Errorf("Expected DefaultTimeout to be 30m, got %v", config.DefaultTimeout)
	}

	if !config.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}

	if config.EnableTracing {
		t.Error("Expected EnableTracing to be false")
	}

	if config.MemoryLimitMB != 512 {
		t.Errorf("Expected MemoryLimitMB to be 512, got %d", config.MemoryLimitMB)
	}
}

func TestNewProcessingEngine(t *testing.T) {
	config := EngineConfig{
		Name:               "test-engine",
		MaxConcurrentItems: 5,
		DefaultTimeout:     10 * time.Minute,
		EnableMetrics:      true,
		EnableTracing:      false,
		MemoryLimitMB:      256,
	}

	engine := NewProcessingEngine(config)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	// Test that metrics are initialized
	metrics := engine.GetMetrics()
	if metrics.ItemsProcessed != 0 {
		t.Errorf("Expected ItemsProcessed to be 0, got %d", metrics.ItemsProcessed)
	}

	if metrics.ItemsSuccessful != 0 {
		t.Errorf("Expected ItemsSuccessful to be 0, got %d", metrics.ItemsSuccessful)
	}

	if metrics.ItemsFailed != 0 {
		t.Errorf("Expected ItemsFailed to be 0, got %d", metrics.ItemsFailed)
	}
}

func TestProcessingEngine_RegisterTransformer(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())
	transformer := &MockTransformer{name: "test-transformer"}

	engine.RegisterTransformer("test", transformer)

	// The engine should accept the transformer without error
	// We can't directly test if it's registered without exposing internals,
	// but we can verify it doesn't panic
}

func TestProcessingEngine_RegisterValidator(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())
	validator := &MockValidator{name: "test-validator"}

	engine.RegisterValidator("test", validator)

	// The engine should accept the validator without error
}

func TestProcessingEngine_AddMiddleware(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())
	middleware := &MockMiddleware{name: "test-middleware"}

	engine.AddMiddleware(middleware)

	// The engine should accept the middleware without error
}

func TestProcessingEngine_Process_InvalidRequest(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())

	// Test with nil request
	_, err := engine.Process(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil request")
	}

	// Test with request missing source
	req := &PipelineRequest{
		Direction: DirectionExport,
		Options:   DefaultPipelineOptions(),
	}
	_, err = engine.Process(context.Background(), req)
	if err == nil {
		t.Error("Expected error for request missing source")
	}

	// Test with request missing destination
	req.Source = &MockSource{}
	_, err = engine.Process(context.Background(), req)
	if err == nil {
		t.Error("Expected error for request missing destination")
	}
}

func TestProcessingEngine_Process_ValidRequest(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())

	// Create mock source and destination
	source := &MockSource{
		items: []DataItemContainer{
			*NewDataItem(ObjectTypePage, "page-1", map[string]interface{}{"title": "Test Page"}),
			*NewDataItem(ObjectTypeBlock, "block-1", map[string]interface{}{"content": "Test Block"}),
		},
	}
	destination := &MockDestination{}

	req := &PipelineRequest{
		Direction:   DirectionExport,
		Source:      source,
		Destination: destination,
		Options:     DefaultPipelineOptions(),
	}

	results, err := engine.Process(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if results == nil {
		t.Fatal("Expected results channel")
	}

	// Collect results
	var collectedResults []PipelineResult
	for result := range results {
		collectedResults = append(collectedResults, result)
	}

	// We should get at least some results (exact count depends on processing phases)
	if len(collectedResults) == 0 {
		t.Error("Expected some results")
	}
}

func TestProcessingEngine_GetMetrics(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())

	metrics := engine.GetMetrics()

	// Test initial metrics
	if metrics.ItemsProcessed != 0 {
		t.Errorf("Expected ItemsProcessed to be 0, got %d", metrics.ItemsProcessed)
	}

	if metrics.ItemsSuccessful != 0 {
		t.Errorf("Expected ItemsSuccessful to be 0, got %d", metrics.ItemsSuccessful)
	}

	if metrics.ItemsFailed != 0 {
		t.Errorf("Expected ItemsFailed to be 0, got %d", metrics.ItemsFailed)
	}

	if metrics.TotalProcessingTime != 0 {
		t.Errorf("Expected TotalProcessingTime to be 0, got %v", metrics.TotalProcessingTime)
	}

	if metrics.AverageProcessingTime != 0 {
		t.Errorf("Expected AverageProcessingTime to be 0, got %v", metrics.AverageProcessingTime)
	}

	if metrics.LastProcessedAt != nil {
		t.Errorf("Expected LastProcessedAt to be nil, got %v", metrics.LastProcessedAt)
	}
}

func TestProcessingEngine_Close(t *testing.T) {
	engine := NewProcessingEngine(DefaultEngineConfig())

	err := engine.Close()
	if err != nil {
		t.Errorf("Expected no error closing engine, got %v", err)
	}
}

func TestWorkerPool(t *testing.T) {
	workerCount := 3
	bufferSize := 10
	pool := NewWorkerPool(workerCount, bufferSize)

	if pool == nil {
		t.Fatal("Expected worker pool to be created")
	}

	// Test submitting work
	var wg sync.WaitGroup
	var results []int
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i // Capture loop variable
		pool.Submit(func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, i)
			mu.Unlock()
		})
	}

	wg.Wait()
	pool.Close()

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestWorkerPool_Close(t *testing.T) {
	pool := NewWorkerPool(2, 5)

	// Submit some work
	workDone := false
	pool.Submit(func() {
		workDone = true
	})

	// Give workers time to complete
	time.Sleep(100 * time.Millisecond)

	// Close the pool
	pool.Close()

	if !workDone {
		t.Error("Expected work to be completed before close")
	}
}

// Mock implementations for testing

type MockTransformer struct {
	name string
}

func (t *MockTransformer) Transform(ctx context.Context, item DataItemContainer) (DataItemContainer, error) {
	// Simple mock transformation - just add a property
	item.SetProperty("transformed_by", t.name)
	return item, nil
}

func (t *MockTransformer) GetName() string {
	return t.name
}

func (t *MockTransformer) GetDescription() string {
	return "Mock transformer for testing"
}

type MockValidator struct {
	name string
}

func (v *MockValidator) Validate(ctx context.Context, item DataItemContainer) ValidationState {
	return ValidationState{
		IsValid:     true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		ValidatedAt: time.Now(),
		Rules:       []string{v.name},
	}
}

func (v *MockValidator) GetName() string {
	return v.name
}

func (v *MockValidator) GetDescription() string {
	return "Mock validator for testing"
}

type MockMiddleware struct {
	name string
}

func (m *MockMiddleware) PreProcess(ctx context.Context, item *DataItemContainer) error {
	item.SetProperty("preprocessed_by", m.name)
	return nil
}

func (m *MockMiddleware) PostProcess(ctx context.Context, item *DataItemContainer) error {
	item.SetProperty("postprocessed_by", m.name)
	return nil
}

func (m *MockMiddleware) GetName() string {
	return m.name
}

type MockSource struct {
	items []DataItemContainer
}

func (s *MockSource) Read(ctx context.Context, req *ReadRequest) (<-chan DataItemContainer, error) {
	results := make(chan DataItemContainer, len(s.items))

	go func() {
		defer close(results)
		for _, item := range s.items {
			select {
			case results <- item:
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

func (s *MockSource) Validate(req *ReadRequest) error {
	return nil
}

func (s *MockSource) SupportsType(objType ObjectType) bool {
	return true
}

func (s *MockSource) Config() SourceConfig {
	return SourceConfig{
		Type: "mock",
		Name: "Mock Source",
	}
}

func (s *MockSource) Close() error {
	return nil
}

type MockDestination struct {
	items []DataItemContainer
	mu    sync.Mutex
}

func (d *MockDestination) Write(ctx context.Context, item DataItemContainer) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items = append(d.items, item)
	return nil
}

func (d *MockDestination) Batch(ctx context.Context, items []DataItemContainer) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items = append(d.items, items...)
	return nil
}

func (d *MockDestination) SupportsType(objType ObjectType) bool {
	return true
}

func (d *MockDestination) Validate(item DataItemContainer) error {
	return nil
}

func (d *MockDestination) Config() DestinationConfig {
	return DestinationConfig{
		Type: "mock",
		Name: "Mock Destination",
	}
}

func (d *MockDestination) Close() error {
	return nil
}

func (d *MockDestination) GetItems() []DataItemContainer {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]DataItemContainer{}, d.items...)
}
