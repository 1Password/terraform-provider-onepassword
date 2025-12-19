package util

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	defaultMaxRetryAttempts = 5
)

// retry retries an operation when it returns a retryable error
// If refreshVersion is provided, it will be called when an error occurs
func retry(ctx context.Context, operation func() error, errorKeywords []string, refreshVersion func() error, errorMsg string, maxAttempts int) error {
	for attempt := range maxAttempts {
		err := operation()
		if err == nil {
			return nil
		}

		// Check if error contains any of the error messages
		errStr := err.Error()
		isRetryable := false
		for _, keyword := range errorKeywords {
			if strings.Contains(errStr, keyword) {
				// If the error contains a retryable keyword, set isRetryable to true
				isRetryable = true
				break
			}
		}

		if !isRetryable {
			return err
		}

		// Retryable error occurred, refresh version if callback provided
		if refreshVersion != nil {
			if refreshErr := refreshVersion(); refreshErr != nil {
				return fmt.Errorf("failed to refresh version after retryable error: %w (original error: %w)", refreshErr, err)
			}
		}

		// Don't sleep on the last attempt
		if attempt < maxAttempts-1 {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			backoff := calculateBackoff(attempt)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("max retry attempts (%d) reached for operation due to %s", maxAttempts, errorMsg)
}

// RetryOnConflict retries an operation when it returns 409 Conflict errors
func RetryOnConflict(ctx context.Context, operation func() error, refreshVersion func() error, maxAttempts ...int) error {
	attempts := defaultMaxRetryAttempts
	if len(maxAttempts) > 0 && maxAttempts[0] > 0 {
		attempts = maxAttempts[0]
	}
	return retry(ctx, operation, []string{"409", "Conflict", "conflict"}, refreshVersion, "409 Conflict errors", attempts)
}

// RetryUntilCondition retries an operation when it returns 404 Not Found errors
func RetryUntilCondition(ctx context.Context, operation func() (bool, error), maxAttempts ...int) error {
	attempts := defaultMaxRetryAttempts
	if len(maxAttempts) > 0 && maxAttempts[0] > 0 {
		attempts = maxAttempts[0]
	}
	var lastErr error

	err := retry(ctx, func() error {
		done, opErr := operation()
		if opErr != nil {
			lastErr = opErr // Save the error for later
			return opErr
		}

		if done {
			// Condition met - success
			return nil
		}
		// Condition not met - return a retryable error
		return fmt.Errorf("condition not met")
	}, []string{"condition not met", "404", "not found"}, nil, "item not found or condition not satisfied", attempts)

	// If all retries failed, return the last error instead of generic message
	if err != nil && lastErr != nil && strings.Contains(err.Error(), "max retry attempts") {
		return fmt.Errorf("%w (after %d retry attempts)", lastErr, attempts)
	}

	return err
}

// calculateBackoff calculates backoff with jitter
func calculateBackoff(attempt int) time.Duration {
	baseDelay := 100 * time.Millisecond
	maxDelay := 500 * time.Millisecond

	exponentialDelay := baseDelay * time.Duration(1<<uint(attempt))

	if exponentialDelay > maxDelay {
		exponentialDelay = maxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(baseDelay)))

	return exponentialDelay + jitter
}
