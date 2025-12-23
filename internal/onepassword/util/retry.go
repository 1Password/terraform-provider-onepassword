package util

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	maxRetryAttempts = 5
)

// retry retries an operation when it returns a retryable error
func retry(ctx context.Context, operation func() error, errorKeywords []string, errorMsg string, maxAttempts int) error {
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
func RetryOnConflict(ctx context.Context, operation func() error) error {
	return retry(ctx, operation, []string{"409", "Conflict", "conflict"}, "409 Conflict errors", maxRetryAttempts)
}

// RetryUntilCondition retries an operation when it returns 404 Not Found errors
func RetryUntilCondition(ctx context.Context, operation func() (bool, error)) error {
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
	}, []string{"condition not met", "404", "not found"}, "item not found or condition not satisfied", maxRetryAttempts)

	// If all retries failed so return the last error instead of generic message
	if err != nil && lastErr != nil && strings.Contains(err.Error(), "max retry attempts") {
		return fmt.Errorf("%w (after %d retry attempts)", lastErr, maxRetryAttempts)
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
