package errors

import (
	"errors"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	now := time.Now()
	err := &RateLimitError{
		Message: "rate limit exceeded",
		Key:     "test_key",
		Limit:   100,
		Reset:   now,
	}

	expectedMsg := "rate limit exceeded: key=test_key, limit=100, reset=" + now.Format(time.RFC3339)
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without key
	err.Key = ""
	expectedMsg = "rate limit exceeded"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIsRateLimitError(t *testing.T) {
	rateLimitErr := &RateLimitError{Message: "test"}
	regularErr := errors.New("regular error")

	if !IsRateLimitError(rateLimitErr) {
		t.Error("IsRateLimitError should return true for RateLimitError")
	}

	if IsRateLimitError(regularErr) {
		t.Error("IsRateLimitError should return false for regular error")
	}

	if IsRateLimitError(nil) {
		t.Error("IsRateLimitError should return false for nil")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Message: "invalid value",
		Field:   "tokens",
		Value:   0,
	}

	expectedMsg := "invalid value: field=tokens, value=0"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without field
	err.Field = ""
	expectedMsg = "invalid value"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without value
	err.Value = nil
	expectedMsg = "invalid value"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIsValidationError(t *testing.T) {
	validationErr := &ValidationError{Message: "test"}
	regularErr := errors.New("regular error")

	if !IsValidationError(validationErr) {
		t.Error("IsValidationError should return true for ValidationError")
	}

	if IsValidationError(regularErr) {
		t.Error("IsValidationError should return false for regular error")
	}

	if IsValidationError(nil) {
		t.Error("IsValidationError should return false for nil")
	}
}

func TestBackendError(t *testing.T) {
	cause := errors.New("connection failed")
	err := &BackendError{
		Message: "backend unavailable",
		Service: "redis",
		Cause:   cause,
	}

	expectedMsg := "backend unavailable: service=redis"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without service
	err.Service = ""
	expectedMsg = "backend unavailable: connection failed"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without cause
	err.Cause = nil
	expectedMsg = "backend unavailable"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIsBackendError(t *testing.T) {
	backendErr := &BackendError{Message: "test"}
	regularErr := errors.New("regular error")

	if !IsBackendError(backendErr) {
		t.Error("IsBackendError should return true for BackendError")
	}

	if IsBackendError(regularErr) {
		t.Error("IsBackendError should return false for regular error")
	}

	if IsBackendError(nil) {
		t.Error("IsBackendError should return false for nil")
	}
}

func TestBackendErrorUnwrap(t *testing.T) {
	cause := errors.New("connection failed")
	err := &BackendError{
		Message: "backend unavailable",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected unwrapped error to be cause, got %v", unwrapped)
	}

	// Test without cause
	err.Cause = nil
	unwrapped = err.Unwrap()
	if unwrapped != nil {
		t.Errorf("expected unwrapped error to be nil, got %v", unwrapped)
	}
}

func TestTimeoutError(t *testing.T) {
	timeout := 5 * time.Second
	err := &TimeoutError{
		Message: "operation timed out",
		Timeout: timeout,
	}

	expectedMsg := "operation timed out: timeout=5s"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test without timeout
	err.Timeout = 0
	expectedMsg = "operation timed out"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIsTimeoutError(t *testing.T) {
	timeoutErr := &TimeoutError{Message: "test"}
	regularErr := errors.New("regular error")

	if !IsTimeoutError(timeoutErr) {
		t.Error("IsTimeoutError should return true for TimeoutError")
	}

	if IsTimeoutError(regularErr) {
		t.Error("IsTimeoutError should return false for regular error")
	}

	if IsTimeoutError(nil) {
		t.Error("IsTimeoutError should return false for nil")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "additional context")

	if wrappedErr == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	expectedMsg := "additional context: original error"
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("expected wrapped error message '%s', got '%s'", expectedMsg, wrappedErr.Error())
	}

	// Test wrapping nil error
	nilWrapped := Wrap(nil, "context")
	if nilWrapped != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrapf(originalErr, "operation %s failed", "test")

	if wrappedErr == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	expectedMsg := "operation test failed: original error"
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("expected wrapped error message '%s', got '%s'", expectedMsg, wrappedErr.Error())
	}

	// Test wrapping nil error
	nilWrapped := Wrapf(nil, "operation %s failed", "test")
	if nilWrapped != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types are properly initialized
	if ErrRateLimitExceeded == nil {
		t.Error("ErrRateLimitExceeded should not be nil")
	}

	if ErrInvalidTokens == nil {
		t.Error("ErrInvalidTokens should not be nil")
	}

	if ErrInvalidKey == nil {
		t.Error("ErrInvalidKey should not be nil")
	}

	if ErrBackendUnavailable == nil {
		t.Error("ErrBackendUnavailable should not be nil")
	}

	if ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}

	// Test error type assertions
	if !IsRateLimitError(ErrRateLimitExceeded) {
		t.Error("ErrRateLimitExceeded should be a RateLimitError")
	}

	if !IsValidationError(ErrInvalidTokens) {
		t.Error("ErrInvalidTokens should be a ValidationError")
	}

	if !IsValidationError(ErrInvalidKey) {
		t.Error("ErrInvalidKey should be a ValidationError")
	}

	if !IsBackendError(ErrBackendUnavailable) {
		t.Error("ErrBackendUnavailable should be a BackendError")
	}

	if !IsTimeoutError(ErrTimeout) {
		t.Error("ErrTimeout should be a TimeoutError")
	}
}

func TestErrorChaining(t *testing.T) {
	// Test error chaining with Wrap
	originalErr := errors.New("database connection failed")
	wrapped1 := Wrap(originalErr, "failed to connect to database")
	wrapped2 := Wrap(wrapped1, "user authentication failed")

	expectedMsg := "user authentication failed: failed to connect to database: database connection failed"
	if wrapped2.Error() != expectedMsg {
		t.Errorf("expected chained error message '%s', got '%s'", expectedMsg, wrapped2.Error())
	}

	// Test error chaining with Wrapf
	wrapped3 := Wrapf(wrapped2, "request %s failed", "GET /api/user")
	expectedMsg = "request GET /api/user failed: user authentication failed: failed to connect to database: database connection failed"
	if wrapped3.Error() != expectedMsg {
		t.Errorf("expected chained error message '%s', got '%s'", expectedMsg, wrapped3.Error())
	}
}
