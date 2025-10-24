package config

import (
	"fmt"
	"os"
)

func GetServiceAccountToken() (string, error) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("OP_SERVICE_ACCOUNT_TOKEN environment variable is required")
	}
	return token, nil
}
