package backend

import (
	"context"
	"time"

	"github.com/devrob-go/go-rate-limiter/pkg/errors"
)

// Backend defines the interface for rate limiter backends
type Backend interface {
	// Take attempts to consume the specified number of tokens from the bucket
	// Returns true if tokens were successfully consumed, false if rate limit exceeded
	// The error is returned if there's a backend failure
	Take(ctx context.Context, key string, tokens int) (bool, error)

	// Reset clears the rate limit for a specific key
	Reset(ctx context.Context, key string) error

	// GetInfo returns information about the current state of a key
	GetInfo(ctx context.Context, key string) (*TokenInfo, error)

	// SetLimit sets a custom limit for a specific key
	SetLimit(ctx context.Context, key string, limit int, refill time.Duration) error

	// Close gracefully shuts down the backend
	Close(ctx context.Context) error

	// HealthCheck performs a health check on the backend
	HealthCheck(ctx context.Context) error
}

// TokenInfo contains information about the current state of a token bucket
type TokenInfo struct {
	Key        string        `json:"key"`
	Tokens     int           `json:"tokens"`
	MaxTokens  int           `json:"max_tokens"`
	RefillRate time.Duration `json:"refill_rate"`
	LastRefill time.Time     `json:"last_refill"`
	NextRefill time.Time     `json:"next_refill"`
	ResetTime  time.Time     `json:"reset_time"`
}

// Options contains configuration options for backends
type Options struct {
	DefaultLimit    int           `json:"default_limit"`
	DefaultRefill   time.Duration `json:"default_refill"`
	DefaultBurst    int           `json:"default_burst"`
	MaxKeys         int           `json:"max_keys"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// DefaultOptions returns default options for backends
func DefaultOptions() *Options {
	return &Options{
		DefaultLimit:    100,
		DefaultRefill:   time.Second,
		DefaultBurst:    10,
		MaxKeys:         10000,
		CleanupInterval: 5 * time.Minute,
	}
}

// Validate validates the options
func (o *Options) Validate() error {
	if o.DefaultLimit <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "default_limit must be positive")
	}

	if o.DefaultRefill <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "default_refill must be positive")
	}

	if o.DefaultBurst <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "default_burst must be positive")
	}

	if o.MaxKeys <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "max_keys must be positive")
	}

	if o.CleanupInterval <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "cleanup_interval must be positive")
	}

	return nil
}

// WithLimit returns new options with custom limit
func (o *Options) WithLimit(limit int) *Options {
	newOpts := *o
	newOpts.DefaultLimit = limit
	return &newOpts
}

// WithRefill returns new options with custom refill rate
func (o *Options) WithRefill(refill time.Duration) *Options {
	newOpts := *o
	newOpts.DefaultRefill = refill
	return &newOpts
}

// WithBurst returns new options with custom burst
func (o *Options) WithBurst(burst int) *Options {
	newOpts := *o
	newOpts.DefaultBurst = burst
	return &newOpts
}
