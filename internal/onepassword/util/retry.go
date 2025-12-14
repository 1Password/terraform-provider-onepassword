package util

import (
	"context"
	"fmt"
	"strings"
)

// RetryOnConflict retries an operation when it returns 409 Conflict errors
func RetryOnConflict(ctx context.Context, operation func() error) error {
	maxAttempts := 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			// Operation succeeded
			return nil
		}

		errStr := err.Error()
		if strings.Contains(errStr, "409") || strings.Contains(errStr, "Conflict") {
			// 409 error - retry
			continue
		}

		// Non-409 error - return immediately
		return err
	}

	return fmt.Errorf("max retry attempts (3) reached for operation due to 409 Conflict errors")
}
