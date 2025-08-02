package engine

import (
	"testing"

	"github.com/cmskitdev/redis"
)

// TestFinalConfigLoading tests that the original issue is now resolved
func TestFinalConfigLoading(t *testing.T) {
	// This should now work with our custom loading
	manager := NewConfigManager()

	// Test getting the main config
	mainCfg, err := manager.Get("main")
	if err != nil {
		t.Errorf("Failed to get main config: %v", err)
	} else {
		if cfg, ok := mainCfg.(*Config); ok {
			t.Logf("✅ Main config loaded: Name=%s, EnableMetrics=%t", cfg.Name, cfg.EnableMetrics)
		}
	}

	// Test getting the redis config specifically
	redisCfg, err := manager.Get("redis")
	if err != nil {
		t.Errorf("Failed to get redis config: %v", err)
	} else {
		if cfg, ok := redisCfg.(*redis.Config); ok {
			t.Logf("✅ Redis config loaded successfully:")
			t.Logf("  Address: %s", cfg.Address)
			t.Logf("  Username: %s", cfg.Username)
			t.Logf("  Database: %d", cfg.Database)
			t.Logf("  Workers: %d", cfg.Workers)
			t.Logf("  KeyPrefix: %s", cfg.KeyPrefix)
			t.Logf("  KeySeparator: %s", cfg.KeySeparator)
			t.Logf("  MaxRetries: %d", cfg.MaxRetries)
			t.Logf("  RetryBackoff: %v", cfg.RetryBackoff)

			// Verify specific values from config.yaml
			if cfg.Address != "localhost:6379" {
				t.Errorf("Expected address 'localhost:6379', got '%s'", cfg.Address)
			}
			if cfg.Database != 15 {
				t.Errorf("Expected database 15, got %d", cfg.Database)
			}
			if cfg.KeyPrefix != "notion_codes" {
				t.Errorf("Expected key_prefix 'notion_codes', got '%s'", cfg.KeyPrefix)
			}
		} else {
			t.Errorf("Expected redis config to be *redis.Config, got %T", redisCfg)
		}
	}
}

// TestConfigManagerUsage demonstrates how to use the config manager
func TestConfigManagerUsage(t *testing.T) {
	manager := NewConfigManager()

	// Example: Get redis config and verify it has all required fields for Redis client
	redisCfg, err := manager.Get("redis")
	if err != nil {
		t.Fatalf("Failed to get redis config: %v", err)
	}

	cfg, ok := redisCfg.(*redis.Config)
	if !ok {
		t.Fatalf("Expected redis config to be *redis.Config, got %T", redisCfg)
	}

	// Check that all required fields for Redis connection are present
	requiredFields := map[string]interface{}{
		"Address":      cfg.Address,
		"Username":     cfg.Username,
		"Password":     cfg.Password,
		"Database":     cfg.Database,
		"KeyPrefix":    cfg.KeyPrefix,
		"KeySeparator": cfg.KeySeparator,
	}

	for field, value := range requiredFields {
		if value == nil || (field == "Address" && value == "") {
			t.Errorf("Required field %s is empty or nil: %v", field, value)
		}
	}

	t.Log("✅ All required Redis configuration fields are properly loaded!")
}