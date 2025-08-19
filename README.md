# Go Rate Limiter

A production-grade, high-performance rate limiting library for Go applications with support for multiple backends and advanced features.

## Features

- **Multiple Backends**: In-memory and Redis backends with consistent API
- **Token Bucket Algorithm**: Efficient rate limiting with configurable refill rates
- **Context Support**: Full context cancellation and timeout support
- **Production Ready**: Comprehensive error handling, health checks, and monitoring
- **Thread Safe**: Concurrent access with proper locking mechanisms
- **Configurable**: Flexible configuration with sensible defaults
- **Comprehensive Testing**: Extensive test coverage with benchmarks
- **Clean API**: Simple, intuitive interface for common use cases

## Installation

```bash
go get github.com/devrob-go/go-rate-limiter
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/devrob-go/go-rate-limiter/pkg/limiter"
    "github.com/devrob-go/go-rate-limiter/pkg/backend"
    "github.com/devrob-go/go-rate-limiter/pkg/config"
)

func main() {
    // Create configuration
    cfg := config.DefaultConfig()
    
    // Create in-memory backend
    backend, err := backend.NewInMemoryBackend(backend.DefaultOptions())
    if err != nil {
        panic(err)
    }
    
    // Create rate limiter
    limiter, err := limiter.New(backend, cfg)
    if err != nil {
        panic(err)
    }
    defer limiter.Close(context.Background())
    
    // Use rate limiter
    ctx := context.Background()
    allowed, err := limiter.Take(ctx, "user_123", 1)
    if err != nil {
        panic(err)
    }
    
    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Rate limit exceeded")
    }
}
```

### Redis Backend

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/devrob-go/go-rate-limiter/pkg/limiter"
    "github.com/devrob-go/go-rate-limiter/pkg/backend"
    "github.com/devrob-go/go-rate-limiter/pkg/config"
)

func main() {
    // Create configuration
    cfg := config.DefaultConfig()
    
    // Create Redis backend
    backend, err := backend.NewRedisBackend("redis://localhost:6379", backend.DefaultOptions())
    if err != nil {
        panic(err)
    }
    
    // Create rate limiter
    limiter, err := limiter.New(backend, cfg)
    if err != nil {
        panic(err)
    }
    defer limiter.Close(context.Background())
    
    // Use rate limiter
    ctx := context.Background()
    allowed, err := limiter.Take(ctx, "user_123", 1)
    if err != nil {
        panic(err)
    }
    
    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Rate limit exceeded")
    }
}
```

### Custom Limits

```go
// Set custom limits for specific keys
err := limiter.TakeWithLimit(ctx, "premium_user", 1, 1000, time.Minute)
if err != nil {
    panic(err)
}

// Check if request would be allowed without consuming tokens
allowed, err := limiter.IsAllowed(ctx, "user_123", 1)
if err != nil {
    panic(err)
}

if allowed {
    fmt.Println("Request would be allowed")
}
```

### Wait for Tokens

```go
// Wait until tokens become available
err := limiter.Wait(ctx, "user_123", 1)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("Timeout waiting for tokens")
    } else {
        panic(err)
    }
}

fmt.Println("Tokens available")
```

### Get Token Information

```go
// Get current token bucket state
info, err := limiter.GetInfo(ctx, "user_123")
if err != nil {
    panic(err)
}

fmt.Printf("Tokens: %d/%d, Next refill: %s\n", 
    info.Tokens, info.MaxTokens, info.NextRefill.Format(time.RFC3339))
```

## Configuration

### Default Configuration

```go
cfg := config.DefaultConfig()
// Default: 100 tokens, 1 second refill, 10 burst, 5 minute cleanup
```

### Custom Configuration

```go
cfg := config.DefaultConfig().
    WithDefaults(50, 2*time.Second, 5).
    WithRedis("redis://localhost:6379").
    WithInMemory(10*time.Minute, 5000)
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `DefaultLimit` | Maximum tokens per bucket | 100 |
| `DefaultRefill` | Token refill rate | 1 second |
| `DefaultBurst` | Burst allowance | 10 |
| `MaxKeys` | Maximum number of keys | 10,000 |
| `CleanupInterval` | Cleanup frequency | 5 minutes |
| `EnableMetrics` | Enable metrics collection | true |
| `EnableLogging` | Enable structured logging | true |

## Backend Options

### In-Memory Backend

```go
options := backend.DefaultOptions().
    WithLimit(200).
    WithRefill(500*time.Millisecond).
    WithBurst(20)

backend, err := backend.NewInMemoryBackend(options)
```

### Redis Backend

```go
options := backend.DefaultOptions().
    WithLimit(200).
    WithRefill(500*time.Millisecond).
    WithBurst(20)

backend, err := backend.NewRedisBackend("redis://localhost:6379", options)
```

## Error Handling

The library provides comprehensive error handling with custom error types:

```go
import "github.com/devrob-go/go-rate-limiter/pkg/errors"

// Check error types
if errors.IsRateLimitExceeded(err) {
    // Handle rate limit exceeded
}

if errors.IsValidationError(err) {
    // Handle validation error
}

if errors.IsBackendError(err) {
    // Handle backend error
}

if errors.IsTimeoutError(err) {
    // Handle timeout error
}
```

## Health Checks

```go
// Check backend health
err := limiter.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
}

// Check limiter health
err = limiter.HealthCheck(ctx)
if err != nil {
    log.Printf("Limiter health check failed: %v", err)
}
```

## Graceful Shutdown

```go
// Gracefully shutdown the rate limiter
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := limiter.Close(ctx)
if err != nil {
    log.Printf("Error during shutdown: %v", err)
}
```

## Testing

### Run Tests

```bash
go test ./...
```

### Run Benchmarks

```bash
go test -bench=. ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

## Performance

The library is designed for high performance:

- **In-Memory Backend**: Sub-millisecond response times
- **Redis Backend**: Optimized with Lua scripts for atomic operations
- **Concurrent Access**: Thread-safe with minimal locking overhead
- **Memory Efficient**: Automatic cleanup of expired buckets

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   RateLimiter   │    │     Backend     │    │   Storage       │
│                 │───▶│   Interface     │───▶│   (Memory/      │
│ - Validation    │    │                 │    │    Redis)       │
│ - Context       │    │ - Take()        │    │                 │
│ - Thread Safety │    │ - Reset()       │    │                 │
│ - Error Handling│    │ - GetInfo()     │    │                 │
└─────────────────┘    │ - SetLimit()    │    └─────────────────┘
                       │ - Close()       │
                       │ - HealthCheck() │
                       └─────────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Token bucket algorithm implementation
- Redis Lua script optimization
- Comprehensive error handling patterns
- Production-ready testing strategies
