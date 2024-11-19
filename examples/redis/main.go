package main

import (
	"fmt"

	"github.com/devrob-go/go-rate-limiter/limiter"
)

func main() {
	backend := limiter.NewRedisBackend("localhost:6379")
	rateLimiter := limiter.New(backend)

	for i := 0; i < 7; i++ {
		ok, _ := rateLimiter.Take("user_123", 1)
		fmt.Printf("Request %d: Allowed = %v\n", i+1, ok)
	}
}
