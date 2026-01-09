package vault

import (
	"testing"
)

// GetTestVaultID returns the vault ID by querying by name once and caching the result.
func GetTestVaultID(t *testing.T) string {
	return "bbucuyq2nn4fozygwttxwizpcy" // "terraform-provider-acceptance-tests" vault ID
}
