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
func retry(ctx context.Context, operation func() error, errorKeywords []string, maxAttempts int) error {
	var err error

	for attempt := range maxAttempts {
		err = operation()
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

	return err
}

// RetryOnConflict retries an operation when it returns 409 Conflict errors
func RetryOnConflict(ctx context.Context, operation func() error) error {
	return retry(ctx, operation, []string{"409", "Conflict", "conflict"}, maxRetryAttempts)
}

// Retry500ForConnectDelete retries a delete operation when it returns 500 Something went wrong only for Connect.
// This is because Connect returns 500 for conflicts.
// This is temporary until Connect is fixed and starts to return 409 for conflicts.
func Retry500ForConnectDelete(ctx context.Context, operation func() error) error {
	return retry(ctx, operation, []string{"500", "Something went wrong"}, maxRetryAttempts)
}

// Retry404UntilCondition retries an operation when it returns 404 Not Found errors
func Retry404UntilCondition(ctx context.Context, operation func() (bool, error)) error {
	return retry(ctx, func() error {
		done, opErr := operation()
		if opErr != nil {
			return opErr
		}
		if done {
			return nil
		}

		// Fallback: callers should return errors with "condition not met" when condition isn't satisfied
		return fmt.Errorf("condition not met")
	}, []string{"condition not met", "404", "not found"}, maxRetryAttempts)
}

// calculateBackoff calculates backoff with jitter
func calculateBackoff(attempt int) time.Duration {
	baseDelay := 100 * time.Millisecond
	maxDelay := 500 * time.Millisecond

	exponentialDelay := baseDelay * time.Duration(1<<uint(attempt))
	jitter := time.Duration(rand.Int63n(int64(baseDelay)))

	delay := exponentialDelay + jitter
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
