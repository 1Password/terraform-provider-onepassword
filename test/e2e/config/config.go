package config

import (
	"fmt"
	"os"
)

// TestConfig holds configuration for e2e tests
type TestConfig struct {
	ServiceAccountToken string
}

// GetTestConfig retrieves test configuration from environment variables
func GetTestConfig() (*TestConfig, error) {
	config := &TestConfig{
		ServiceAccountToken: os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"),
	}

	if config.ServiceAccountToken == "" {
		return nil, fmt.Errorf("OP_SERVICE_ACCOUNT_TOKEN environment variable is required")
	}

	return config, nil
}
