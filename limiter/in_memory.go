package limiter

import (
	"sync"
	"time"
)

type inMemoryBackend struct {
	store         sync.Map
	defaultBucket bucket
}

type bucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewInMemoryBackend initializes an in-memory rate limiter backend.
func NewInMemoryBackend(maxTokens int, refillRate time.Duration) Backend {
	return &inMemoryBackend{
		store: sync.Map{},
		defaultBucket: bucket{
			maxTokens:  maxTokens,
			refillRate: refillRate,
			lastRefill: time.Now(),
		},
	}
}

func (b *inMemoryBackend) Take(key string, tokens int) (bool, error) {
	now := time.Now()
	val, _ := b.store.LoadOrStore(key, &bucket{
		tokens:     b.defaultBucket.maxTokens,
		maxTokens:  b.defaultBucket.maxTokens,
		refillRate: b.defaultBucket.refillRate,
		lastRefill: now,
	})
	bkt := val.(*bucket)

	// Refill tokens
	elapsed := now.Sub(bkt.lastRefill)
	newTokens := int(elapsed / bkt.refillRate)
	if newTokens > 0 {
		bkt.tokens = min(bkt.maxTokens, bkt.tokens+newTokens)
		bkt.lastRefill = now
	}

	// Consume tokens
	if bkt.tokens >= tokens {
		bkt.tokens -= tokens
		return true, nil
	}
	return false, nil

}

func (b *inMemoryBackend) Reset(key string) error {
	b.store.Delete(key)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
