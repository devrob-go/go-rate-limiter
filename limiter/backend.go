// limiter/backend.go
package limiter

type Backend interface {
	Take(key string, tokens int) (bool, error) // Check if tokens are available and consume them
	Reset(key string) error                   // Reset the limit for a specific key
}
