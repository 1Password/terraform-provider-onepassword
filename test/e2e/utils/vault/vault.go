package vault

import (
	"sync"
	"testing"
)

var (
	testVaultIDOnce sync.Once
	testVaultID     string
)

// GetTestVaultID returns the vault ID by querying by name once and caching the result.
func GetTestVaultID(t *testing.T) string {
	return "terraform-provider-acceptance-tests"
}
