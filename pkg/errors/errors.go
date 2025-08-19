package errors

import (
	"fmt"
	"time"
)

// Error types for the rate limiter package
var (
	ErrRateLimitExceeded  = &RateLimitError{Message: "rate limit exceeded"}
	ErrInvalidTokens      = &ValidationError{Message: "invalid number of tokens"}
	ErrInvalidKey         = &ValidationError{Message: "invalid key provided"}
	ErrBackendUnavailable = &BackendError{Message: "backend service unavailable"}
	ErrTimeout            = &TimeoutError{Message: "operation timed out"}
)

// RateLimitError represents an error when the rate limit is exceeded
type RateLimitError struct {
	Message string
	Key     string
	Limit   int
	Reset   time.Time
}

func (e *RateLimitError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("%s: key=%s, limit=%d, reset=%s", e.Message, e.Key, e.Limit, e.Reset.Format(time.RFC3339))
	}
	return e.Message
}

// IsRateLimitError checks if the error is a RateLimitError
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// ValidationError represents validation errors
type ValidationError struct {
	Message string
	Field   string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: field=%s, value=%v", e.Message, e.Field, e.Value)
	}
	return e.Message
}

// IsValidationError checks if the error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// BackendError represents backend service errors
type BackendError struct {
	Message string
	Service string
	Cause   error
}

func (e *BackendError) Error() string {
	if e.Service != "" {
		return fmt.Sprintf("%s: service=%s", e.Message, e.Service)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause error
func (e *BackendError) Unwrap() error {
	return e.Cause
}

// IsBackendError checks if the error is a BackendError
func IsBackendError(err error) bool {
	_, ok := err.(*BackendError)
	return ok
}

// TimeoutError represents timeout errors
type TimeoutError struct {
	Message string
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	if e.Timeout > 0 {
		return fmt.Sprintf("%s: timeout=%v", e.Message, e.Timeout)
	}
	return e.Message
}

// IsTimeoutError checks if the error is a TimeoutError
func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted additional context
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
