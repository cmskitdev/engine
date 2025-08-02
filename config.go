package engine

import (
	"fmt"
	"time"

	"github.com/cmskitdev/redis"
	"github.com/mateothegreat/go-config/config"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/validation"
)

// GetDefaultTimeout parses and returns the DefaultTimeout as time.Duration
func (c *Config) GetDefaultTimeout() (time.Duration, error) {
	return time.ParseDuration(c.DefaultTimeout)
}

type RedisConfig struct {
	// From plugins.Config
	Reporter        bool          `json:"reporter" yaml:"reporter"`
	RuntimeTimeout  time.Duration `json:"runtime_timeout" yaml:"runtime_timeout"`
	RequestDelay    time.Duration `json:"request_delay" yaml:"request_delay"`
	ContinueOnError bool          `json:"continue_on_error" yaml:"continue_on_error"`

	// Redis connection settings (from ClientConfig)
	Address  string `json:"address" yaml:"address"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Database int    `json:"database" yaml:"database"`

	// Key configuration (from ClientConfig)
	KeyPrefix    string `json:"key_prefix" yaml:"key_prefix"`
	KeySeparator string `json:"key_separator" yaml:"key_separator"`

	// Data settings (from ClientConfig)
	TTL         time.Duration `json:"ttl" yaml:"ttl"`
	PrettyJSON  bool          `json:"pretty_json" yaml:"pretty_json"`
	IncludeMeta bool          `json:"include_meta" yaml:"include_meta"`

	// Performance settings (from ClientConfig)
	Pipeline     bool          `json:"pipeline" yaml:"pipeline"`
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	RetryBackoff time.Duration `json:"retry_backoff" yaml:"retry_backoff"`

	// Redis plugin specific
	Workers int  `json:"workers" yaml:"workers"`
	Flush   bool `json:"flush" yaml:"flush"`
}

// Plugin is t
// Config configures the processing engine.
type Config struct {
	Name               string `json:"name"`
	MaxConcurrentItems int    `json:"max_concurrent_items"`
	DefaultTimeout     string `json:"default_timeout" yaml:"default_timeout"`
	EnableMetrics      bool   `json:"enable_metrics"`
	EnableTracing      bool   `json:"enable_tracing"`
	MemoryLimitMB      int    `json:"memory_limit_mb"`

	Redis *RedisConfig `json:"redis" yaml:"redis"`
}

type ConfigManager struct {
	configs map[string]any
}

func NewConfigManager() *ConfigManager {
	cfg := &Config{
		Redis: &RedisConfig{}, // Initialize the Redis field
	}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		config.FromEnv(sources.EnvOpts{Prefix: "SIMPLE"}),
	).WithValidationStrategy(validation.StrategyAuto).Build(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to load configuration: %v", err)
	// }

	fmt.Printf("err: %v\n", err)
	if err != nil {
		fmt.Printf("error details: %+v\n", err)
	}

	fmt.Println(cfg)

	return &ConfigManager{
		configs: map[string]any{
			"redis": &redis.Config{},
		},
	}
}

func (c *ConfigManager) Get(name string) (any, error) {
	cfg, ok := c.configs[name]
	if !ok {
		return nil, fmt.Errorf("config %s not found", name)
	}
	return cfg, nil
}

func (c *ConfigManager) Set(name string, cfg any) {
	c.configs[name] = cfg
}

// DefaultConfig returns sensible defaults for the engine.
func DefaultConfig() Config {
	return Config{
		Name:               "default-engine",
		MaxConcurrentItems: 10,
		DefaultTimeout:     "30m",
		EnableMetrics:      true,
		EnableTracing:      false,
		MemoryLimitMB:      512,
	}
}
