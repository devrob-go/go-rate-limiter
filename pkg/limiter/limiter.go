package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/devrob-go/go-rate-limiter/pkg/backend"
	"github.com/devrob-go/go-rate-limiter/pkg/config"
	"github.com/devrob-go/go-rate-limiter/pkg/errors"
)

// RateLimiter provides rate limiting functionality with configurable backends
type RateLimiter struct {
	backend backend.Backend
	config  *config.Config
	mu      sync.RWMutex
	closed  bool
}

// New creates a new rate limiter with the given backend and configuration
func New(backend backend.Backend, cfg *config.Config) (*RateLimiter, error) {
	if backend == nil {
		return nil, errors.Wrap(errors.ErrBackendUnavailable, "backend cannot be nil")
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &RateLimiter{
		backend: backend,
		config:  cfg,
	}, nil
}

// Take attempts to consume the specified number of tokens from the bucket
// Returns true if tokens were successfully consumed, false if rate limit exceeded
func (r *RateLimiter) Take(ctx context.Context, key string, tokens int) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return false, errors.Wrap(errors.ErrBackendUnavailable, "rate limiter is closed")
	}

	if err := r.validateKey(key); err != nil {
		return false, err
	}

	if err := r.validateTokens(tokens); err != nil {
		return false, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	// Attempt to take tokens from the backend
	allowed, err := r.backend.Take(ctx, key, tokens)
	if err != nil {
		return false, errors.Wrap(err, "failed to take tokens from backend")
	}

	return allowed, nil
}

// TakeWithLimit attempts to consume tokens with a custom limit for the key
func (r *RateLimiter) TakeWithLimit(ctx context.Context, key string, tokens int, limit int, refill time.Duration) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return false, errors.Wrap(errors.ErrBackendUnavailable, "rate limiter is closed")
	}

	if err := r.validateKey(key); err != nil {
		return false, err
	}

	if err := r.validateTokens(tokens); err != nil {
		return false, err
	}

	if limit <= 0 {
		return false, errors.Wrap(errors.ErrInvalidTokens, "limit must be positive")
	}

	if refill <= 0 {
		return false, errors.Wrap(errors.ErrInvalidTokens, "refill rate must be positive")
	}

	// Set custom limit for this key
	if err := r.backend.SetLimit(ctx, key, limit, refill); err != nil {
		return false, errors.Wrap(err, "failed to set custom limit")
	}

	// Attempt to take tokens
	return r.backend.Take(ctx, key, tokens)
}

// Reset clears the rate limit for a specific key
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "rate limiter is closed")
	}

	if err := r.validateKey(key); err != nil {
		return err
	}

	return r.backend.Reset(ctx, key)
}

// GetInfo returns information about the current state of a key
func (r *RateLimiter) GetInfo(ctx context.Context, key string) (*backend.TokenInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, errors.Wrap(errors.ErrBackendUnavailable, "rate limiter is closed")
	}

	if err := r.validateKey(key); err != nil {
		return nil, err
	}

	return r.backend.GetInfo(ctx, key)
}

// IsAllowed checks if a request would be allowed without consuming tokens
func (r *RateLimiter) IsAllowed(ctx context.Context, key string, tokens int) (bool, error) {
	info, err := r.GetInfo(ctx, key)
	if err != nil {
		return false, err
	}

	return info.Tokens >= tokens, nil
}

// Wait waits until tokens become available or context is cancelled
func (r *RateLimiter) Wait(ctx context.Context, key string, tokens int) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context cancelled while waiting")
		case <-ticker.C:
			allowed, err := r.IsAllowed(ctx, key, tokens)
			if err != nil {
				return err
			}
			if allowed {
				return nil
			}
		}
	}
}

// Close gracefully shuts down the rate limiter
func (r *RateLimiter) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if err := r.backend.Close(ctx); err != nil {
		return errors.Wrap(err, "failed to close backend")
	}

	return nil
}

// HealthCheck performs a health check on the rate limiter
func (r *RateLimiter) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "rate limiter is closed")
	}

	return r.backend.HealthCheck(ctx)
}

// GetConfig returns a copy of the current configuration
func (r *RateLimiter) GetConfig() *config.Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.config == nil {
		return nil
	}

	// Return a copy to prevent external modification
	configCopy := *r.config
	return &configCopy
}

// validateKey validates the key parameter
func (r *RateLimiter) validateKey(key string) error {
	if key == "" {
		return errors.Wrap(errors.ErrInvalidKey, "key cannot be empty")
	}

	if len(key) > 256 {
		return errors.Wrap(errors.ErrInvalidKey, "key too long (max 256 characters)")
	}

	return nil
}

// validateTokens validates the tokens parameter
func (r *RateLimiter) validateTokens(tokens int) error {
	if tokens <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "tokens must be positive")
	}

	if tokens > r.config.DefaultLimit*10 {
		return errors.Wrap(errors.ErrInvalidTokens, "tokens exceed reasonable limit")
	}

	return nil
}

// String returns a string representation of the rate limiter
func (r *RateLimiter) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return "RateLimiter{closed=true}"
	}

	return fmt.Sprintf("RateLimiter{backend=%T, config=%+v}", r.backend, r.config)
}
