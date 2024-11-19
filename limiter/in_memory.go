// limiter/in_memory.go
package limiter

import (
	"sync"
	"time"
)

type inMemoryBackend struct {
	store sync.Map
}

type bucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewInMemoryBackend initializes an in-memory rate limiter backend.
func NewInMemoryBackend(maxTokens int, refillRate time.Duration) Backend {
	return &inMemoryBackend{}
}

func (b *inMemoryBackend) Take(key string, tokens int) (bool, error) {
	now := time.Now()

	val, _ := b.store.LoadOrStore(key, &bucket{
		tokens:     tokens,
		maxTokens:  tokens,
		refillRate: time.Second,
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
