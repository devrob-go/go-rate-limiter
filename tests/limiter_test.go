// tests/limiter_test.go
package tests

import (
	"testing"
	"time"

	"github.com/devrob-go/go-rate-limiter/limiter"
)

func TestInMemoryBackend(t *testing.T) {
	backend := limiter.NewInMemoryBackend(5, time.Second)
	rateLimiter := limiter.New(backend)

	for i := 0; i < 5; i++ {
		if ok, _ := rateLimiter.Take("test", 1); !ok {
			t.Fatalf("Expected request to succeed, got failure at %d", i)
		}
	}

	if ok, _ := rateLimiter.Take("test", 1); ok {
		t.Fatalf("Expected request to fail after exceeding limit")
	}

	time.Sleep(time.Second)
	if ok, _ := rateLimiter.Take("test", 1); !ok {
		t.Fatalf("Expected request to succeed after refill")
	}
}
