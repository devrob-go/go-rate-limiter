package benchmarks

import (
	"testing"
	"time"

	"github.com/devrob-go/go-rate-limiter/limiter"
)

func BenchmarkInMemoryBackend(b *testing.B) {
	backend := limiter.NewInMemoryBackend(100, time.Millisecond)
	rateLimiter := limiter.New(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimiter.Take("test", 1)
	}
}
