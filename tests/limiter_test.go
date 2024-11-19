package tests

import (
	"testing"
	"time"

	"github.com/devrob-go/go-rate-limiter/limiter"
)

func TestInMemoryBackend(t *testing.T) {
	backend := limiter.NewInMemoryBackend(5, time.Second) // 5 tokens, 1 second refill
	rateLimiter := limiter.New(backend)

	for i := 0; i < 5; i++ {
		if ok, _ := rateLimiter.Take("test", 1); !ok {
			t.Fatalf("Expected request to succeed, got failure at %d", i+1)
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
