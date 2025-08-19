package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultLimit != 100 {
		t.Errorf("expected DefaultLimit to be 100, got %d", config.DefaultLimit)
	}

	if config.DefaultRefill != time.Second {
		t.Errorf("expected DefaultRefill to be 1s, got %v", config.DefaultRefill)
	}

	if config.DefaultBurst != 10 {
		t.Errorf("expected DefaultBurst to be 10, got %d", config.DefaultBurst)
	}

	if config.CleanupInterval != 5*time.Minute {
		t.Errorf("expected CleanupInterval to be 5m, got %v", config.CleanupInterval)
	}

	if config.MaxKeys != 10000 {
		t.Errorf("expected MaxKeys to be 10000, got %d", config.MaxKeys)
	}

	if !config.EnableMetrics {
		t.Error("expected EnableMetrics to be true")
	}

	if !config.EnableLogging {
		t.Error("expected EnableLogging to be true")
	}

	// Test Redis config defaults
	if config.Redis.Addr != "localhost:6379" {
		t.Errorf("expected Redis.Addr to be 'localhost:6379', got %s", config.Redis.Addr)
	}

	if config.Redis.PoolSize != 10 {
		t.Errorf("expected Redis.PoolSize to be 10, got %d", config.Redis.PoolSize)
	}

	if config.Redis.MinIdleConns != 5 {
		t.Errorf("expected Redis.MinIdleConns to be 5, got %d", config.Redis.MinIdleConns)
	}

	if config.Redis.MaxRetries != 3 {
		t.Errorf("expected Redis.MaxRetries to be 3, got %d", config.Redis.MaxRetries)
	}

	if config.Redis.Timeout != 5*time.Second {
		t.Errorf("expected Redis.Timeout to be 5s, got %v", config.Redis.Timeout)
	}

	if config.Redis.DialTimeout != 5*time.Second {
		t.Errorf("expected Redis.DialTimeout to be 5s, got %v", config.Redis.DialTimeout)
	}

	// Test InMemory config defaults
	if config.InMemory.CleanupInterval != 5*time.Minute {
		t.Errorf("expected InMemory.CleanupInterval to be 5m, got %v", config.InMemory.CleanupInterval)
	}

	if config.InMemory.MaxKeys != 10000 {
		t.Errorf("expected InMemory.MaxKeys to be 10000, got %d", config.InMemory.MaxKeys)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid default limit",
			config: &Config{
				DefaultLimit:    0,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				CleanupInterval: 5 * time.Minute,
				MaxKeys:         10000,
			},
			expectError: true,
		},
		{
			name: "invalid default refill",
			config: &Config{
				DefaultLimit:    100,
				DefaultRefill:   0,
				DefaultBurst:    10,
				CleanupInterval: 5 * time.Minute,
				MaxKeys:         10000,
			},
			expectError: true,
		},
		{
			name: "invalid default burst",
			config: &Config{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    0,
				CleanupInterval: 5 * time.Minute,
				MaxKeys:         10000,
			},
			expectError: true,
		},
		{
			name: "invalid cleanup interval",
			config: &Config{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				CleanupInterval: 0,
				MaxKeys:         10000,
			},
			expectError: true,
		},
		{
			name: "invalid max keys",
			config: &Config{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				CleanupInterval: 5 * time.Minute,
				MaxKeys:         0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigWithRedis(t *testing.T) {
	config := DefaultConfig()
	newConfig := config.WithRedis("redis://localhost:6380")

	if newConfig.Redis.Addr != "redis://localhost:6380" {
		t.Errorf("expected Redis.Addr to be 'redis://localhost:6380', got %s", newConfig.Redis.Addr)
	}

	// Original config should remain unchanged
	if config.Redis.Addr != "localhost:6379" {
		t.Errorf("original Redis.Addr should remain 'localhost:6379', got %s", config.Redis.Addr)
	}
}

func TestConfigWithInMemory(t *testing.T) {
	config := DefaultConfig()
	newCleanupInterval := 10 * time.Minute
	newMaxKeys := 5000
	newConfig := config.WithInMemory(newCleanupInterval, newMaxKeys)

	if newConfig.InMemory.CleanupInterval != newCleanupInterval {
		t.Errorf("expected InMemory.CleanupInterval to be %v, got %v", newCleanupInterval, newConfig.InMemory.CleanupInterval)
	}

	if newConfig.InMemory.MaxKeys != newMaxKeys {
		t.Errorf("expected InMemory.MaxKeys to be %d, got %d", newMaxKeys, newConfig.InMemory.MaxKeys)
	}

	// Original config should remain unchanged
	if config.InMemory.CleanupInterval != 5*time.Minute {
		t.Errorf("original InMemory.CleanupInterval should remain 5m, got %v", config.InMemory.CleanupInterval)
	}

	if config.InMemory.MaxKeys != 10000 {
		t.Errorf("original InMemory.MaxKeys should remain 10000, got %d", config.InMemory.MaxKeys)
	}
}

func TestConfigWithDefaults(t *testing.T) {
	config := DefaultConfig()
	newLimit := 200
	newRefill := 2 * time.Second
	newBurst := 20
	newConfig := config.WithDefaults(newLimit, newRefill, newBurst)

	if newConfig.DefaultLimit != newLimit {
		t.Errorf("expected DefaultLimit to be %d, got %d", newLimit, newConfig.DefaultLimit)
	}

	if newConfig.DefaultRefill != newRefill {
		t.Errorf("expected DefaultRefill to be %v, got %v", newRefill, newConfig.DefaultRefill)
	}

	if newConfig.DefaultBurst != newBurst {
		t.Errorf("expected DefaultBurst to be %d, got %d", newBurst, newConfig.DefaultBurst)
	}

	// Original config should remain unchanged
	if config.DefaultLimit != 100 {
		t.Errorf("original DefaultLimit should remain 100, got %d", config.DefaultLimit)
	}

	if config.DefaultRefill != time.Second {
		t.Errorf("original DefaultRefill should remain 1s, got %v", config.DefaultRefill)
	}

	if config.DefaultBurst != 10 {
		t.Errorf("original DefaultBurst should remain 10, got %d", config.DefaultBurst)
	}
}

func TestRedisConfig(t *testing.T) {
	config := DefaultConfig()

	// Test Redis config fields
	if config.Redis.Password != "" {
		t.Errorf("expected Redis.Password to be empty, got %s", config.Redis.Password)
	}

	if config.Redis.DB != 0 {
		t.Errorf("expected Redis.DB to be 0, got %d", config.Redis.DB)
	}
}

func TestInMemoryConfig(t *testing.T) {
	config := DefaultConfig()

	// Test InMemory config fields
	if config.InMemory.CleanupInterval != 5*time.Minute {
		t.Errorf("expected InMemory.CleanupInterval to be 5m, got %v", config.InMemory.CleanupInterval)
	}

	if config.InMemory.MaxKeys != 10000 {
		t.Errorf("expected InMemory.MaxKeys to be 10000, got %d", config.InMemory.MaxKeys)
	}
}

func TestConfigImmutability(t *testing.T) {
	config := DefaultConfig()
	originalLimit := config.DefaultLimit
	originalRefill := config.DefaultRefill

	// Create new configs
	config1 := config.WithDefaults(200, 2*time.Second, 20)
	config2 := config.WithDefaults(300, 3*time.Second, 30)

	// Verify original config is unchanged
	if config.DefaultLimit != originalLimit {
		t.Errorf("original config was modified: expected %d, got %d", originalLimit, config.DefaultLimit)
	}

	if config.DefaultRefill != originalRefill {
		t.Errorf("original config was modified: expected %v, got %v", originalRefill, config.DefaultRefill)
	}

	// Verify new configs have different values
	if config1.DefaultLimit == config2.DefaultLimit {
		t.Error("config1 and config2 should have different DefaultLimit values")
	}

	if config1.DefaultRefill == config2.DefaultRefill {
		t.Error("config1 and config2 should have different DefaultRefill values")
	}
}
