package backend

import (
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.DefaultLimit != 100 {
		t.Errorf("expected DefaultLimit to be 100, got %d", opts.DefaultLimit)
	}

	if opts.DefaultRefill != time.Second {
		t.Errorf("expected DefaultRefill to be 1s, got %v", opts.DefaultRefill)
	}

	if opts.DefaultBurst != 10 {
		t.Errorf("expected DefaultBurst to be 10, got %d", opts.DefaultBurst)
	}

	if opts.MaxKeys != 10000 {
		t.Errorf("expected MaxKeys to be 10000, got %d", opts.MaxKeys)
	}

	if opts.CleanupInterval != 5*time.Minute {
		t.Errorf("expected CleanupInterval to be 5m, got %v", opts.CleanupInterval)
	}
}

func TestOptionsValidation(t *testing.T) {
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
			name: "invalid default limit",
			options: &Options{
				DefaultLimit:    0,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				MaxKeys:         10000,
				CleanupInterval: 5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "invalid default refill",
			options: &Options{
				DefaultLimit:    100,
				DefaultRefill:   0,
				DefaultBurst:    10,
				MaxKeys:         10000,
				CleanupInterval: 5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "invalid default burst",
			options: &Options{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    0,
				MaxKeys:         10000,
				CleanupInterval: 5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "invalid max keys",
			options: &Options{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				MaxKeys:         0,
				CleanupInterval: 5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "invalid cleanup interval",
			options: &Options{
				DefaultLimit:    100,
				DefaultRefill:   time.Second,
				DefaultBurst:    10,
				MaxKeys:         10000,
				CleanupInterval: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

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

func TestOptionsWithLimit(t *testing.T) {
	opts := DefaultOptions()
	newOpts := opts.WithLimit(200)

	if newOpts.DefaultLimit != 200 {
		t.Errorf("expected DefaultLimit to be 200, got %d", newOpts.DefaultLimit)
	}

	// Original options should remain unchanged
	if opts.DefaultLimit != 100 {
		t.Errorf("original DefaultLimit should remain 100, got %d", opts.DefaultLimit)
	}
}

func TestOptionsWithRefill(t *testing.T) {
	opts := DefaultOptions()
	newRefill := 2 * time.Second
	newOpts := opts.WithRefill(newRefill)

	if newOpts.DefaultRefill != newRefill {
		t.Errorf("expected DefaultRefill to be %v, got %v", newRefill, newOpts.DefaultRefill)
	}

	// Original options should remain unchanged
	if opts.DefaultRefill != time.Second {
		t.Errorf("original DefaultRefill should remain 1s, got %v", opts.DefaultRefill)
	}
}

func TestOptionsWithBurst(t *testing.T) {
	opts := DefaultOptions()
	newOpts := opts.WithBurst(20)

	if newOpts.DefaultBurst != 20 {
		t.Errorf("expected DefaultBurst to be 20, got %d", newOpts.DefaultBurst)
	}

	// Original options should remain unchanged
	if opts.DefaultBurst != 10 {
		t.Errorf("original DefaultBurst should remain 10, got %d", opts.DefaultBurst)
	}
}

func TestTokenInfo(t *testing.T) {
	now := time.Now()
	info := &TokenInfo{
		Key:        "test_key",
		Tokens:     50,
		MaxTokens:  100,
		RefillRate: time.Second,
		LastRefill: now,
		NextRefill: now.Add(time.Second),
		ResetTime:  now.Add(time.Second),
	}

	if info.Key != "test_key" {
		t.Errorf("expected Key to be 'test_key', got %s", info.Key)
	}

	if info.Tokens != 50 {
		t.Errorf("expected Tokens to be 50, got %d", info.Tokens)
	}

	if info.MaxTokens != 100 {
		t.Errorf("expected MaxTokens to be 100, got %d", info.MaxTokens)
	}

	if info.RefillRate != time.Second {
		t.Errorf("expected RefillRate to be 1s, got %v", info.RefillRate)
	}
}
