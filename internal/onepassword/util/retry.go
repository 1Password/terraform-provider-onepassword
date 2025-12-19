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
// If refreshVersion is provided, it will be called when a retryable error occurs
func retry(ctx context.Context, operation func() error, errorKeywords []string, refreshVersion func() error, errorMsg string) error {
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		// Check if error contains any of the error messages
		errStr := err.Error()
		isRetryable := false
		for _, keyword := range errorKeywords {
			if strings.Contains(errStr, keyword) {
				isRetryable = true
				break
			}
		}

		if !isRetryable {
			return err
		}

		// Retryable error - refresh version if callback provided
		if refreshVersion != nil {
			if refreshErr := refreshVersion(); refreshErr != nil {
				return fmt.Errorf("failed to refresh version after retryable error: %w (original error: %w)", refreshErr, err)
			}
		}

		// Don't sleep on the last attempt
		if attempt < maxRetryAttempts-1 {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			backoff := calculateBackoff(attempt)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("max retry attempts (%d) reached for operation due to %s", maxRetryAttempts, errorMsg)
}

// RetryOnConflict retries an operation when it returns 409 Conflict errors.
// If refreshVersion is provided, it will be called on 409 errors to fetch and update
// the latest vault version before retrying.
func RetryOnConflict(ctx context.Context, operation func() error, refreshVersion func() error) error {
	return retry(ctx, operation, []string{"409", "Conflict", "conflict"}, refreshVersion, "409 Conflict errors")
}

// RetryUntilCondition retries an operation until a condition is met.
func RetryUntilCondition(ctx context.Context, operation func() (bool, error)) error {
	// Use retry with a wrapper that converts condition check to error-based
	return retry(ctx, func() error {
		done, err := operation()
		if err != nil {
			return err // Non-retryable error
		}
		if done {
			return nil // Condition met - success
		}
		// Condition not met - return a retryable error
		return fmt.Errorf("condition not met")
	}, []string{"condition not met"}, nil, "condition to be met")
}

// calculateBackoff calculates backoff with jitter
func calculateBackoff(attempt int) time.Duration {
	baseDelay := 100 * time.Millisecond
	jitter := time.Duration(rand.Int63n(int64(baseDelay)))
	return baseDelay + jitter
}
