package limiter

type RateLimiter struct {
	backend Backend
}

// New creates a new rate limiter with the given backend.
func New(backend Backend) *RateLimiter {
	return &RateLimiter{backend: backend}
}

// Take attempts to consume the given number of tokens.
func (r *RateLimiter) Take(key string, tokens int) (bool, error) {
	return r.backend.Take(key, tokens)
}

// Reset clears the rate limit for a specific key.
func (r *RateLimiter) Reset(key string) error {
	return r.backend.Reset(key)
}
