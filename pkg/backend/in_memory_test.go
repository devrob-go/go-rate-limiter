package backend

import (
	"context"
	"testing"
	"time"
)

func TestNewInMemoryBackend(t *testing.T) {
	tests := []struct {
		name        string
		options     *Options
		expectError bool
	}{
		{
			name:        "valid options",
			options:     DefaultOptions(),
			expectError: false,
		},
		{
			name:        "nil options uses defaults",
			options:     nil,
			expectError: false,
		},
		{
			name: "invalid options",
			options: &Options{
				DefaultLimit:    0,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				MaxKeys:         10000,
				CleanupInterval: 5 * time.Minute,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewInMemoryBackend(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if backend == nil {
				t.Error("expected backend, got nil")
			}
		})
	}
}

func TestInMemoryBackendTake(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Test successful token consumption
	allowed, err := backend.Take(ctx, "test_key", 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed")
	}

	// Test consuming more tokens than available
	allowed, err = backend.Take(ctx, "test_key", 200)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be denied")
	}

	// Test invalid key
	_, err = backend.Take(ctx, "", 1)
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test invalid tokens
	_, err = backend.Take(ctx, "test_key", 0)
	if err == nil {
		t.Error("expected error for zero tokens")
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = backend.Take(cancelledCtx, "test_key", 1)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestInMemoryBackendReset(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Test successful reset
	err = backend.Reset(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test empty key
	err = backend.Reset(ctx, "")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestInMemoryBackendGetInfo(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Test getting info for new key
	info, err := backend.GetInfo(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info == nil {
		t.Error("expected info, got nil")
	}

	info, err = backend.GetInfo(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info == nil {
		t.Error("expected info, got nil")
		return
	}

	if info.Key != "test_key" {
		t.Errorf("expected key %s, got %s", "test_key", info.Key)
	}

	if info.Tokens != 100 {
		t.Errorf("expected tokens 100, got %d", info.Tokens)
	}

	// Test getting info after consuming tokens
	backend.Take(ctx, "test_key", 10)
	info, err = backend.GetInfo(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info.Tokens != 90 {
		t.Errorf("expected tokens 90, got %d", info.Tokens)
	}

	// Test invalid key
	_, err = backend.GetInfo(ctx, "")
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = backend.GetInfo(cancelledCtx, "test_key")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestInMemoryBackendSetLimit(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Test setting custom limit
	err = backend.SetLimit(ctx, "test_key", 50, 2*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify the limit was set
	info, err := backend.GetInfo(ctx, "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info.MaxTokens != 50 {
		t.Errorf("expected MaxTokens 50, got %d", info.MaxTokens)
	}

	if info.RefillRate != 2*time.Second {
		t.Errorf("expected RefillRate 2s, got %v", info.RefillRate)
	}

	// Test invalid limit
	err = backend.SetLimit(ctx, "test_key", 0, time.Second)
	if err == nil {
		t.Error("expected error for zero limit")
	}

	// Test invalid refill rate
	err = backend.SetLimit(ctx, "test_key", 50, 0)
	if err == nil {
		t.Error("expected error for zero refill rate")
	}

	// Test invalid key
	err = backend.SetLimit(ctx, "", 50, time.Second)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestInMemoryBackendClose(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	ctx := context.Background()

	// Test close
	err = backend.Close(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test operations after close
	_, err = backend.Take(ctx, "test_key", 1)
	if err == nil {
		t.Error("expected error after close")
	}

	// Test multiple close calls
	err = backend.Close(ctx)
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestInMemoryBackendHealthCheck(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Test health check
	err = backend.HealthCheck(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test health check after close
	backend.Close(ctx)
	err = backend.HealthCheck(ctx)
	if err == nil {
		t.Error("expected error after close")
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = backend.HealthCheck(cancelledCtx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestInMemoryBackendTokenRefill(t *testing.T) {
	opts := DefaultOptions()
	opts.DefaultRefill = 100 * time.Millisecond
	backend, err := NewInMemoryBackend(opts)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()

	// Consume all tokens
	allowed, err := backend.Take(ctx, "test_key", 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed")
	}

	// Try to consume more tokens (should fail)
	allowed, err = backend.Take(ctx, "test_key", 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be denied")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should be able to consume tokens again
	allowed, err = backend.Take(ctx, "test_key", 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed after refill")
	}
}

func TestInMemoryBackendConcurrency(t *testing.T) {
	backend, err := NewInMemoryBackend(DefaultOptions())
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Close(context.Background())

	ctx := context.Background()
	key := "concurrent_key"
	numGoroutines := 10
	results := make(chan bool, numGoroutines)

	// Start multiple goroutines to consume tokens concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			allowed, _ := backend.Take(ctx, key, 10)
			results <- allowed
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			allowedCount++
		}
	}

	// Should allow exactly 10 requests (100 tokens / 10 tokens per request)
	if allowedCount != 10 {
		t.Errorf("expected 10 allowed requests, got %d", allowedCount)
	}
}

func TestInMemoryBackendCleanup(t *testing.T) {
	opts := DefaultOptions()
	opts.CleanupInterval = 100 * time.Millisecond
	backend, err := NewInMemoryBackend(opts)
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}

	ctx := context.Background()

	// Create some buckets
	backend.Take(ctx, "key1", 1)
	backend.Take(ctx, "key2", 1)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Close backend to stop cleanup goroutine
	backend.Close(ctx)
}
