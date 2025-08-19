package config

import (
	"fmt"
	"time"
)

// Config holds the configuration for the rate limiter
type Config struct {
	// General settings
	DefaultLimit  int           `json:"default_limit" yaml:"default_limit"`
	DefaultRefill time.Duration `json:"default_refill" yaml:"default_refill"`
	DefaultBurst  int           `json:"default_burst" yaml:"default_burst"`

	// Redis settings
	Redis RedisConfig `json:"redis" yaml:"redis"`

	// In-memory settings
	InMemory InMemoryConfig `json:"in_memory" yaml:"in_memory"`

	// Performance settings
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	MaxKeys         int           `json:"max_keys" yaml:"max_keys"`

	// Monitoring settings
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
	EnableLogging bool `json:"enable_logging" yaml:"enable_logging"`
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addr         string        `json:"addr" yaml:"addr"`
	Password     string        `json:"password" yaml:"password"`
	DB           int           `json:"db" yaml:"db"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
}

// InMemoryConfig holds in-memory backend configuration
type InMemoryConfig struct {
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	MaxKeys         int           `json:"max_keys" yaml:"max_keys"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultLimit:    100,
		DefaultRefill:   time.Second,
		DefaultBurst:    10,
		CleanupInterval: 5 * time.Minute,
		MaxKeys:         10000,
		EnableMetrics:   true,
		EnableLogging:   true,
		Redis: RedisConfig{
			Addr:         "localhost:6379",
			PoolSize:     10,
			MinIdleConns: 5,
			MaxRetries:   3,
			Timeout:      5 * time.Second,
			DialTimeout:  5 * time.Second,
		},
		InMemory: InMemoryConfig{
			CleanupInterval: 5 * time.Minute,
			MaxKeys:         10000,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultLimit <= 0 {
		return fmt.Errorf("default_limit must be positive, got %d", c.DefaultLimit)
	}

	if c.DefaultRefill <= 0 {
		return fmt.Errorf("default_refill must be positive, got %v", c.DefaultRefill)
	}

	if c.DefaultBurst <= 0 {
		return fmt.Errorf("default_burst must be positive, got %d", c.DefaultBurst)
	}

	if c.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive, got %v", c.CleanupInterval)
	}

	if c.MaxKeys <= 0 {
		return fmt.Errorf("max_keys must be positive, got %d", c.MaxKeys)
	}

	return nil
}

// WithRedis returns a new config with Redis settings
func (c *Config) WithRedis(addr string) *Config {
	newConfig := *c
	newConfig.Redis.Addr = addr
	return &newConfig
}

// WithInMemory returns a new config with in-memory settings
func (c *Config) WithInMemory(cleanupInterval time.Duration, maxKeys int) *Config {
	newConfig := *c
	newConfig.InMemory.CleanupInterval = cleanupInterval
	newConfig.InMemory.MaxKeys = maxKeys
	return &newConfig
}

// WithDefaults returns a new config with custom defaults
func (c *Config) WithDefaults(limit int, refill time.Duration, burst int) *Config {
	newConfig := *c
	newConfig.DefaultLimit = limit
	newConfig.DefaultRefill = refill
	newConfig.DefaultBurst = burst
	return &newConfig
}
