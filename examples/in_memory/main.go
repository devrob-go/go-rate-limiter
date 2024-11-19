package main

import (
	"fmt"
	"time"

	"github.com/devrob-go/go-rate-limiter/limiter"
)

func main() {
	backend := limiter.NewInMemoryBackend(5, time.Second)
	rateLimiter := limiter.New(backend)

	for i := 0; i < 7; i++ {
		ok, _ := rateLimiter.Take("user_123", 1)
		fmt.Printf("Request %d: Allowed = %v\n", i+1, ok)
		time.Sleep(200 * time.Millisecond)
	}
}
