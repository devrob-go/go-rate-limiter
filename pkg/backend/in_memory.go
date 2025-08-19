package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/devrob-go/go-rate-limiter/pkg/errors"
)

// inMemoryBackend provides an in-memory implementation of the Backend interface
// It uses a token bucket algorithm with configurable limits and refill rates
type inMemoryBackend struct {
	store         sync.Map
	options       *Options
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	mu            sync.RWMutex
	closed        bool
}

// bucket represents a token bucket for rate limiting
type bucket struct {
	Key        string        `json:"key"`
	Tokens     int           `json:"tokens"`
	MaxTokens  int           `json:"max_tokens"`
	RefillRate time.Duration `json:"refill_rate"`
	LastRefill time.Time     `json:"last_refill"`
	NextRefill time.Time     `json:"next_refill"`
	ResetTime  time.Time     `json:"reset_time"`
	mu         sync.RWMutex
}

// NewInMemoryBackend creates a new in-memory backend with the given options
func NewInMemoryBackend(options *Options) (Backend, error) {
	if options == nil {
		options = DefaultOptions()
	}

	if err := options.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	backend := &inMemoryBackend{
		options:       options,
		cleanupTicker: time.NewTicker(options.CleanupInterval),
		stopCleanup:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go backend.cleanupRoutine()

	return backend, nil
}

// Take attempts to consume tokens from the bucket
func (b *inMemoryBackend) Take(ctx context.Context, key string, tokens int) (bool, error) {
	if b.closed {
		return false, errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	if err := validateKey(key); err != nil {
		return false, err
	}

	if err := validateTokens(tokens); err != nil {
		return false, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	// Get or create bucket
	bkt := b.getOrCreateBucket(key)

	// Refill tokens based on time elapsed
	bkt.refillTokens()

	// Check if we have enough tokens
	bkt.mu.Lock()
	defer bkt.mu.Unlock()

	if bkt.Tokens >= tokens {
		bkt.Tokens -= tokens
		return true, nil
	}

	return false, nil
}

// Reset clears the rate limit for a specific key
func (b *inMemoryBackend) Reset(ctx context.Context, key string) error {
	if b.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	if err := validateKey(key); err != nil {
		return err
	}

	b.store.Delete(key)
	return nil
}

// GetInfo returns information about the current state of a key
func (b *inMemoryBackend) GetInfo(ctx context.Context, key string) (*TokenInfo, error) {
	if b.closed {
		return nil, errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	if err := validateKey(key); err != nil {
		return nil, err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	bkt := b.getOrCreateBucket(key)
	bkt.refillTokens()

	bkt.mu.RLock()
	defer bkt.mu.RUnlock()

	return &TokenInfo{
		Key:        bkt.Key,
		Tokens:     bkt.Tokens,
		MaxTokens:  bkt.MaxTokens,
		RefillRate: bkt.RefillRate,
		LastRefill: bkt.LastRefill,
		NextRefill: bkt.NextRefill,
		ResetTime:  bkt.ResetTime,
	}, nil
}

// SetLimit sets a custom limit for a specific key
func (b *inMemoryBackend) SetLimit(ctx context.Context, key string, limit int, refill time.Duration) error {
	if b.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	if err := validateKey(key); err != nil {
		return err
	}

	if limit <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "limit must be positive")
	}

	if refill <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "refill rate must be positive")
	}

	bkt := b.getOrCreateBucket(key)

	bkt.mu.Lock()
	defer bkt.mu.Unlock()

	bkt.MaxTokens = limit
	bkt.RefillRate = refill
	bkt.ResetTime = time.Now().Add(refill)

	return nil
}

// Close gracefully shuts down the backend
func (b *inMemoryBackend) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Stop cleanup goroutine
	close(b.stopCleanup)
	if b.cleanupTicker != nil {
		b.cleanupTicker.Stop()
	}

	return nil
}

// HealthCheck performs a health check on the backend
func (b *inMemoryBackend) HealthCheck(ctx context.Context) error {
	if b.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	// Simple health check - try to access the store
	b.store.Range(func(key, value interface{}) bool {
		return false // Stop after first iteration
	})

	return nil
}

// getOrCreateBucket gets an existing bucket or creates a new one
func (b *inMemoryBackend) getOrCreateBucket(key string) *bucket {
	val, loaded := b.store.Load(key)
	if loaded {
		return val.(*bucket)
	}

	// Create new bucket
	now := time.Now()
	newBucket := &bucket{
		Key:        key,
		Tokens:     b.options.DefaultLimit,
		MaxTokens:  b.options.DefaultLimit,
		RefillRate: b.options.DefaultRefill,
		LastRefill: now,
		NextRefill: now.Add(b.options.DefaultRefill),
		ResetTime:  now.Add(b.options.DefaultRefill),
	}

	// Store the bucket
	b.store.Store(key, newBucket)
	return newBucket
}

// refillTokens refills tokens based on time elapsed since last refill
func (bkt *bucket) refillTokens() {
	bkt.mu.Lock()
	defer bkt.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bkt.LastRefill)

	// Calculate how many tokens to add
	tokensToAdd := int(elapsed / bkt.RefillRate)

	if tokensToAdd > 0 {
		// Add tokens, but don't exceed max
		bkt.Tokens = min(bkt.MaxTokens, bkt.Tokens+tokensToAdd)
		bkt.LastRefill = now
		bkt.NextRefill = now.Add(bkt.RefillRate)
		bkt.ResetTime = now.Add(bkt.RefillRate)
	}
}

// cleanupRoutine periodically cleans up expired buckets
func (b *inMemoryBackend) cleanupRoutine() {
	for {
		select {
		case <-b.cleanupTicker.C:
			b.cleanupExpiredBuckets()
		case <-b.stopCleanup:
			return
		}
	}
}

// cleanupExpiredBuckets removes buckets that haven't been used recently
func (b *inMemoryBackend) cleanupExpiredBuckets() {
	cutoff := time.Now().Add(-b.options.CleanupInterval * 2)

	b.store.Range(func(key, value interface{}) bool {
		bkt := value.(*bucket)

		bkt.mu.RLock()
		lastUsed := bkt.LastRefill
		bkt.mu.RUnlock()

		if lastUsed.Before(cutoff) {
			b.store.Delete(key)
		}

		return true
	})
}

// validateKey validates the key parameter
func validateKey(key string) error {
	if key == "" {
		return errors.Wrap(errors.ErrInvalidKey, "key cannot be empty")
	}

	if len(key) > 256 {
		return errors.Wrap(errors.ErrInvalidKey, "key too long (max 256 characters)")
	}

	return nil
}

// validateTokens validates the tokens parameter
func validateTokens(tokens int) error {
	if tokens <= 0 {
		return errors.Wrap(errors.ErrInvalidTokens, "tokens must be positive")
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// String returns a string representation of the backend
func (b *inMemoryBackend) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return "InMemoryBackend{closed=true}"
	}

	return fmt.Sprintf("InMemoryBackend{options=%+v}", b.options)
}
