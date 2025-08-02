package engine

import (
	"testing"

	"github.com/cmskitdev/plugins"
)

func TestEnginePluginRegistration(t *testing.T) {
	engine := NewEngine[any](NewEngineOptions{
		Plugins: []string{"redis"},
	})

	// Test that plugin registry has the plugin
	plugin, exists := engine.pluginRegistry.Get("redis")
	if !exists {
		t.Fatal("Plugin should be registered in registry")
	}

	if plugin.ID() != "redis" {
		t.Errorf("Expected plugin ID 'redis', got '%s'", plugin.ID())
	}

	// Test that plugin has handlers
	handlers := plugin.Handlers()
	if len(handlers) == 0 {
		t.Error("Plugin should have handlers")
	}

	// Test that init handler exists
	if _, exists := handlers[plugins.EventInit]; !exists {
		t.Error("Plugin should have init handler")
	}
}
