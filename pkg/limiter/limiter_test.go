package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/devrob-go/go-rate-limiter/pkg/backend"
	"github.com/devrob-go/go-rate-limiter/pkg/config"
)

// mockBackend is a mock implementation of the Backend interface for testing
type mockBackend struct {
	takeFunc     func(ctx context.Context, key string, tokens int) (bool, error)
	resetFunc    func(ctx context.Context, key string) error
	getInfoFunc  func(ctx context.Context, key string) (*backend.TokenInfo, error)
	setLimitFunc func(ctx context.Context, key string, limit int, refill time.Duration) error
	closeFunc    func(ctx context.Context) error
	healthFunc   func(ctx context.Context) error
}

func (m *mockBackend) Take(ctx context.Context, key string, tokens int) (bool, error) {
	if m.takeFunc != nil {
		return m.takeFunc(ctx, key, tokens)
	}
	return true, nil
}

func (m *mockBackend) Reset(ctx context.Context, key string) error {
	if m.resetFunc != nil {
		return m.resetFunc(ctx, key)
	}
	return nil
}

func (m *mockBackend) GetInfo(ctx context.Context, key string) (*backend.TokenInfo, error) {
	if m.getInfoFunc != nil {
		return m.getInfoFunc(ctx, key)
	}
	return &backend.TokenInfo{
		Key:        key,
		Tokens:     100,
		MaxTokens:  100,
		RefillRate: time.Second,
		LastRefill: time.Now(),
		NextRefill: time.Now().Add(time.Second),
		ResetTime:  time.Now().Add(time.Second),
	}, nil
}

func (m *mockBackend) SetLimit(ctx context.Context, key string, limit int, refill time.Duration) error {
	if m.setLimitFunc != nil {
		return m.setLimitFunc(ctx, key, limit, refill)
	}
	return nil
}

func (m *mockBackend) Close(ctx context.Context) error {
	_ = ctx // Use context parameter to avoid linter warning
	if m.closeFunc != nil {
		return m.closeFunc(ctx)
	}
	return nil
}

func (m *mockBackend) HealthCheck(ctx context.Context) error {
	if m.healthFunc != nil {
		return m.healthFunc(ctx)
	}
	return nil
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		backend     backend.Backend
		config      *config.Config
		expectError bool
	}{
		{
			name:        "valid backend and config",
			backend:     &mockBackend{},
			config:      config.DefaultConfig(),
			expectError: false,
		},
		{
			name:        "nil backend",
			backend:     nil,
			config:      config.DefaultConfig(),
			expectError: true,
		},
		{
			name:        "nil config uses default",
			backend:     &mockBackend{},
			config:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := New(tt.backend, tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if limiter == nil {
				t.Error("expected limiter, got nil")
			}
		})
	}
}

func TestTake(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test successful token consumption
	allowed, err := limiter.Take(ctx, "test_key", 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed")
	}

	// Test invalid key
	_, err = limiter.Take(ctx, "", 1)
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test invalid tokens
	_, err = limiter.Take(ctx, "test_key", 0)
	if err == nil {
		t.Error("expected error for zero tokens")
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = limiter.Take(cancelledCtx, "test_key", 1)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestTakeWithLimit(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{
		setLimitFunc: func(ctx context.Context, key string, limit int, refill time.Duration) error {
			return nil
		},
		takeFunc: func(ctx context.Context, key string, tokens int) (bool, error) {
			return true, nil
		},
	}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test successful token consumption with custom limit
	allowed, err := limiter.TakeWithLimit(ctx, "test_key", 1, 50, 2*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed")
	}

	// Test invalid limit
	_, err = limiter.TakeWithLimit(ctx, "test_key", 1, 0, time.Second)
	if err == nil {
		t.Error("expected error for zero limit")
	}

	// Test invalid refill rate
	_, err = limiter.TakeWithLimit(ctx, "test_key", 1, 50, 0)
	if err == nil {
		t.Error("expected error for zero refill rate")
	}

	// Test invalid key
	_, err = limiter.TakeWithLimit(ctx, "", 1, 50, time.Second)
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test invalid tokens
	_, err = limiter.TakeWithLimit(ctx, "test_key", 0, 50, time.Second)
	if err == nil {
		t.Error("expected error for zero tokens")
	}
}

func TestReset(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test successful reset
	err = limiter.Reset(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test empty key
	err = limiter.Reset(ctx, "")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestGetInfo(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	info, err := limiter.GetInfo(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info == nil {
		t.Error("expected info, got nil")
	}

	if info.Key != "test_key" {
		t.Errorf("expected key %s, got %s", "test_key", info.Key)
	}

	// Test empty key
	_, err = limiter.GetInfo(ctx, "")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestIsAllowed(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{
		getInfoFunc: func(ctx context.Context, key string) (*backend.TokenInfo, error) {
			return &backend.TokenInfo{
				Key:        key,
				Tokens:     50,
				MaxTokens:  100,
				RefillRate: time.Second,
				LastRefill: time.Now(),
				NextRefill: time.Now().Add(time.Second),
				ResetTime:  time.Now().Add(time.Second),
			}, nil
		},
	}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test allowed request
	allowed, err := limiter.IsAllowed(ctx, "test_key", 25)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed")
	}

	// Test denied request
	allowed, err = limiter.IsAllowed(ctx, "test_key", 75)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be denied")
	}

	// Test empty key
	_, err = limiter.IsAllowed(ctx, "", 25)
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test zero tokens (this should be allowed since 50 >= 0)
	allowed, err = limiter.IsAllowed(ctx, "test_key", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request with zero tokens to be allowed")
	}
}

func TestWait(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{
		getInfoFunc: func(ctx context.Context, key string) (*backend.TokenInfo, error) {
			return &backend.TokenInfo{
				Key:        key,
				Tokens:     0,
				MaxTokens:  100,
				RefillRate: time.Second,
				LastRefill: time.Now(),
				NextRefill: time.Now().Add(time.Second),
				ResetTime:  time.Now().Add(time.Second),
			}, nil
		},
	}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = limiter.Wait(cancelledCtx, "test_key", 1)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Test empty key
	err = limiter.Wait(ctx, "", 1)
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test zero tokens (this should not cause an error in Wait)
	err = limiter.Wait(ctx, "test_key", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test close
	err = limiter.Close(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test operations after close
	_, err = limiter.Take(ctx, "test_key", 1)
	if err == nil {
		t.Error("expected error after close")
	}

	// Test multiple close calls
	err = limiter.Close(ctx)
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestHealthCheck(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test health check
	err = limiter.HealthCheck(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test health check after close
	limiter.Close(ctx)
	err = limiter.HealthCheck(ctx)
	if err == nil {
		t.Error("expected error after close")
	}
}

func TestGetConfig(t *testing.T) {
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test get config
	retrievedConfig := limiter.GetConfig()
	if retrievedConfig == nil {
		t.Error("expected config, got nil")
	}

	if retrievedConfig.DefaultLimit != config.DefaultLimit {
		t.Errorf("expected DefaultLimit %d, got %d", config.DefaultLimit, retrievedConfig.DefaultLimit)
	}

	// Test that returned config is a copy
	retrievedConfig.DefaultLimit = 999
	if limiter.GetConfig().DefaultLimit == 999 {
		t.Error("modifying returned config should not affect original")
	}
}

func TestString(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test string representation
	str := limiter.String()
	if str == "" {
		t.Error("expected non-empty string representation")
	}

	// Test string representation after close
	limiter.Close(ctx)
	str = limiter.String()
	if str == "" {
		t.Error("expected non-empty string representation after close")
	}
}

func TestValidation(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}

	// Test key validation
	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{"valid key", "test_key", false},
		{"empty key", "", true},
		{"key too long", string(make([]byte, 257)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := limiter.Take(ctx, tt.key, 1)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	// Test token validation
	tests2 := []struct {
		name        string
		tokens      int
		expectError bool
	}{
		{"valid tokens", 1, false},
		{"zero tokens", 0, true},
		{"negative tokens", -1, true},
		{"tokens exceed limit", 1001, true}, // 100 * 10 + 1
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			_, err := limiter.Take(ctx, "test_key", tt.tokens)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}
	config := config.DefaultConfig()

	limiter, err := New(backend, config)
	if err != nil {
		t.Fatalf("failed to create limiter: %v", err)
	}
	defer limiter.Close(ctx)

	// Test concurrent operations
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := limiter.Take(ctx, "concurrent_key", 1)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("unexpected error in goroutine: %v", err)
		}
	}
}
