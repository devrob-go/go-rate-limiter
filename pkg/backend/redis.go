package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/devrob-go/go-rate-limiter/pkg/errors"
	"github.com/go-redis/redis/v8"
)

// redisBackend provides a Redis implementation of the Backend interface
// It uses Lua scripts for atomic operations and supports connection pooling
type redisBackend struct {
	client  *redis.Client
	options *Options
	closed  bool
}

// NewRedisBackend creates a new Redis backend with the given Redis URL and options
func NewRedisBackend(redisURL string, options *Options) (Backend, error) {
	if redisURL == "" {
		return nil, errors.Wrap(errors.ErrBackendUnavailable, "Redis URL cannot be empty")
	}

	if options == nil {
		options = DefaultOptions()
	}

	if err := options.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid options")
	}

	// Parse Redis URL and create client
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Redis URL")
	}

	// Override with custom options if provided
	if options.DefaultLimit > 0 {
		// Use default limit for token bucket
	}
	if options.DefaultRefill > 0 {
		// Use default refill rate for token bucket
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, errors.Wrap(err, "failed to connect to Redis")
	}

	return &redisBackend{
		client:  client,
		options: options,
	}, nil
}

// Take attempts to consume tokens from the bucket using a Lua script
func (r *redisBackend) Take(ctx context.Context, key string, tokens int) (bool, error) {
	if r.closed {
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

	// Use Lua script for atomic token consumption
	script := `
		local key = KEYS[1]
		local tokens_to_consume = tonumber(ARGV[1])
		local max_tokens = tonumber(ARGV[2])
		local refill_rate = tonumber(ARGV[3])
		local current_time = tonumber(ARGV[4])
		
		-- Get current bucket state
		local bucket_data = redis.call('HMGET', key, 'tokens', 'max_tokens', 'refill_rate', 'last_refill')
		local current_tokens = tonumber(bucket_data[1]) or max_tokens
		local bucket_max_tokens = tonumber(bucket_data[2]) or max_tokens
		local bucket_refill_rate = tonumber(bucket_data[3]) or refill_rate
		local last_refill = tonumber(bucket_data[4]) or current_time
		
		-- Calculate refill
		local time_elapsed = current_time - last_refill
		local tokens_to_add = math.floor(time_elapsed / bucket_refill_rate)
		
		if tokens_to_add > 0 then
			current_tokens = math.min(bucket_max_tokens, current_tokens + tokens_to_add)
			last_refill = current_time
		end
		
		-- Check if we can consume tokens
		if current_tokens >= tokens_to_consume then
			current_tokens = current_tokens - tokens_to_consume
			
			-- Update bucket state
			redis.call('HMSET', key, 
				'tokens', current_tokens,
				'max_tokens', bucket_max_tokens,
				'refill_rate', bucket_refill_rate,
				'last_refill', last_refill,
				'updated_at', current_time
			)
			
			-- Set expiration (cleanup after 24 hours of inactivity)
			redis.call('EXPIRE', key, 86400)
			
			return 1
		else
			return 0
		end
	`

	// Execute Lua script
	currentTime := time.Now().Unix()
	result, err := r.client.Eval(ctx, script, []string{key}, tokens, r.options.DefaultLimit, r.options.DefaultRefill.Milliseconds(), currentTime).Int()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to execute Redis script")
	}

	return result == 1, nil
}

// Reset clears the rate limit for a specific key
func (r *redisBackend) Reset(ctx context.Context, key string) error {
	if r.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	if err := validateKey(key); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(err, "failed to delete Redis key")
	}

	return nil
}

// GetInfo returns information about the current state of a key
func (r *redisBackend) GetInfo(ctx context.Context, key string) (*TokenInfo, error) {
	if r.closed {
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

	// Get bucket data from Redis
	bucketData, err := r.client.HMGet(ctx, key, "tokens", "max_tokens", "refill_rate", "last_refill", "updated_at").Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist, return default info
			return &TokenInfo{
				Key:        key,
				Tokens:     r.options.DefaultLimit,
				MaxTokens:  r.options.DefaultLimit,
				RefillRate: r.options.DefaultRefill,
				LastRefill: time.Now(),
				NextRefill: time.Now().Add(r.options.DefaultRefill),
				ResetTime:  time.Now().Add(r.options.DefaultRefill),
			}, nil
		}
		return nil, errors.Wrap(err, "failed to get bucket info from Redis")
	}

	// Parse bucket data
	var tokens, maxTokens int
	var refillRate time.Duration
	var lastRefill time.Time

	if bucketData[0] != nil {
		if t, ok := bucketData[0].(string); ok {
			fmt.Sscanf(t, "%d", &tokens)
		}
	}

	if bucketData[1] != nil {
		if t, ok := bucketData[1].(string); ok {
			fmt.Sscanf(t, "%d", &maxTokens)
		}
	}

	if bucketData[2] != nil {
		if t, ok := bucketData[2].(string); ok {
			if ms, err := time.ParseDuration(t + "ms"); err == nil {
				refillRate = ms
			}
		}
	}

	if bucketData[3] != nil {
		if t, ok := bucketData[3].(string); ok {
			if ts, err := time.Parse(time.RFC3339, t); err == nil {
				lastRefill = ts
			}
		}
	}

	// Skip updated_at field for now

	// Use defaults if values are missing
	if maxTokens == 0 {
		maxTokens = r.options.DefaultLimit
	}
	if refillRate == 0 {
		refillRate = r.options.DefaultRefill
	}
	if lastRefill.IsZero() {
		lastRefill = time.Now()
	}

	// Calculate next refill and reset time
	nextRefill := lastRefill.Add(refillRate)
	resetTime := lastRefill.Add(refillRate)

	return &TokenInfo{
		Key:        key,
		Tokens:     tokens,
		MaxTokens:  maxTokens,
		RefillRate: refillRate,
		LastRefill: lastRefill,
		NextRefill: nextRefill,
		ResetTime:  resetTime,
	}, nil
}

// SetLimit sets a custom limit for a specific key
func (r *redisBackend) SetLimit(ctx context.Context, key string, limit int, refill time.Duration) error {
	if r.closed {
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

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	// Update bucket limits in Redis
	now := time.Now()
	err := r.client.HMSet(ctx, key,
		"max_tokens", limit,
		"refill_rate", refill.Milliseconds(),
		"last_refill", now.Format(time.RFC3339),
		"updated_at", now.Format(time.RFC3339),
	).Err()

	if err != nil {
		return errors.Wrap(err, "failed to set bucket limits in Redis")
	}

	// Set expiration
	if err := r.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		return errors.Wrap(err, "failed to set key expiration")
	}

	return nil
}

// Close gracefully shuts down the backend
func (r *redisBackend) Close(ctx context.Context) error {
	if r.closed {
		return nil
	}

	r.closed = true

	if r.client != nil {
		return r.client.Close()
	}

	return nil
}

// HealthCheck performs a health check on the backend
func (r *redisBackend) HealthCheck(ctx context.Context) error {
	if r.closed {
		return errors.Wrap(errors.ErrBackendUnavailable, "backend is closed")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context cancelled")
	default:
	}

	// Simple ping to Redis
	if err := r.client.Ping(ctx).Err(); err != nil {
		return errors.Wrap(err, "Redis health check failed")
	}

	return nil
}

// String returns a string representation of the backend
func (r *redisBackend) String() string {
	if r.closed {
		return "RedisBackend{closed=true}"
	}

	return fmt.Sprintf("RedisBackend{client=%T, options=%+v}", r.client, r.options)
}
