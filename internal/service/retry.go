package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/disk"
)

// RetryConfig defines retry behavior for transient errors.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig returns sensible defaults: 3 attempts, 100ms base, 2s max.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// RetryableError wraps an error to indicate it should be retried.
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError creates a new RetryableError.
func NewRetryableError(err error) *RetryableError {
	return &RetryableError{Err: err}
}

// isTransientError checks if an error is transient and should be retried.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation - not transient
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for DiskError with rate limiting or server errors
	var diskErr *disk.DiskError
	if errors.As(err, &diskErr) {
		if diskErr.IsRateLimited() {
			return true
		}
		// 5xx server errors are transient
		if diskErr.StatusCode >= 500 && diskErr.StatusCode < 600 {
			return true
		}
		return false
	}

	// Check for RetryableError wrapper
	var retryable *RetryableError
	if errors.As(err, &retryable) {
		return true
	}

	return false
}

// DoWithRetry executes the given function, retrying on transient errors.
// Uses exponential backoff: delay = min(baseDelay * 2^attempt, maxDelay)
func DoWithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// If not transient, fail immediately
		if !isTransientError(err) {
			return err
		}

		// If this was the last attempt, return the error
		if attempt == config.MaxAttempts-1 {
			return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Calculate backoff delay
		delay := time.Duration(math.Min(
			float64(config.BaseDelay)*math.Pow(2, float64(attempt)),
			float64(config.MaxDelay),
		))

		// Wait with context support
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
		}
	}

	return lastErr
}
