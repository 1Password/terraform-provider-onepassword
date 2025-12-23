package util

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRetryUntilCondition(t *testing.T) {
	tests := map[string]struct {
		operation       func() (bool, error)
		expectedErr     string
		expectedRetries int
	}{
		"should succeed immediately when condition is met": {
			operation: func() (bool, error) {
				return true, nil
			},
			expectedErr:     "",
			expectedRetries: 1,
		},
		"should succeed after retrying on 404": {
			operation: func() func() (bool, error) {
				attempt := 0
				return func() (bool, error) {
					attempt++
					if attempt < 3 {
						return false, errors.New("status 404: item not found")
					}
					return true, nil
				}
			}(),
			expectedErr:     "",
			expectedRetries: 3,
		},
		"should return wrapped 404 error after max retries": {
			operation: func() (bool, error) {
				return false, errors.New("status 404: item not found")
			},
			expectedErr:     "status 404: item not found (after 5 retry attempts)",
			expectedRetries: 5,
		},
		"should return non-retryable error immediately": {
			operation: func() (bool, error) {
				return false, errors.New("status 500: internal server error")
			},
			expectedErr:     "status 500: internal server error",
			expectedRetries: 1,
		},
		"should retry on condition not met": {
			operation: func() func() (bool, error) {
				attempt := 0
				return func() (bool, error) {
					attempt++
					if attempt < 3 {
						return false, nil // Condition not met, no error
					}
					return true, nil
				}
			}(),
			expectedErr:     "",
			expectedRetries: 3,
		},
		"should return generic error when condition not met and no error after max retries": {
			operation: func() (bool, error) {
				return false, nil // Condition not met, no error
			},
			expectedErr:     "max retry attempts (5) reached for operation due to item not found or condition not satisfied",
			expectedRetries: 5,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			attempts := 0
			originalOp := test.operation

			// Wrap operation to count attempts
			operation := func() (bool, error) {
				attempts++
				return originalOp()
			}

			err := RetryUntilCondition(context.Background(), operation)

			// Check error
			if test.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", test.expectedErr)
				} else if !strings.Contains(err.Error(), test.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", test.expectedErr, err)
				}
			}

			// Check retry attempts
			if attempts != test.expectedRetries {
				t.Errorf("Expected %d retry attempts, got %d", test.expectedRetries, attempts)
			}
		})
	}
}

func TestRetryOnConflict(t *testing.T) {
	tests := map[string]struct {
		operation       func() error
		refreshVersion  func() error
		expectedErr     string
		expectedRetries int
		expectedRefresh int
	}{
		"should succeed immediately when no error": {
			operation: func() error {
				return nil
			},
			refreshVersion:  nil,
			expectedErr:     "",
			expectedRetries: 1,
			expectedRefresh: 0,
		},
		"should succeed after retrying on 409": {
			operation: func() func() error {
				attempt := 0
				return func() error {
					attempt++
					if attempt < 3 {
						return errors.New("status 409: Conflict")
					}
					return nil
				}
			}(),
			refreshVersion:  nil,
			expectedErr:     "",
			expectedRetries: 3,
			expectedRefresh: 0,
		},
		"should return generic error after max retries on 409": {
			operation: func() error {
				return errors.New("status 409: Conflict")
			},
			refreshVersion:  nil,
			expectedErr:     "max retry attempts (5) reached for operation due to 409 Conflict errors",
			expectedRetries: 5,
			expectedRefresh: 0,
		},
		"should return non-retryable error immediately": {
			operation: func() error {
				return errors.New("status 500: internal server error")
			},
			refreshVersion:  nil,
			expectedErr:     "status 500: internal server error",
			expectedRetries: 1,
			expectedRefresh: 0,
		},

		"should retry on different 409 error formats": {
			operation: func() func() error {
				attempt := 0
				return func() error {
					attempt++
					if attempt < 2 {
						return errors.New("conflict error")
					}
					return nil
				}
			}(),
			refreshVersion:  nil,
			expectedErr:     "",
			expectedRetries: 2,
			expectedRefresh: 0,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			attempts := 0
			refreshCalls := 0
			originalOp := test.operation

			// Wrap operation to count attempts
			operation := func() error {
				attempts++
				return originalOp()
			}

			err := RetryOnConflict(context.Background(), operation)

			// Check error
			if test.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", test.expectedErr)
				} else if !strings.Contains(err.Error(), test.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", test.expectedErr, err)
				}
			}

			// Check retry attempts
			if attempts != test.expectedRetries {
				t.Errorf("Expected %d retry attempts, got %d", test.expectedRetries, attempts)
			}

			// Check refreshVersion calls
			if refreshCalls != test.expectedRefresh {
				t.Errorf("Expected %d refreshVersion calls, got %d", test.expectedRefresh, refreshCalls)
			}
		})
	}
}
