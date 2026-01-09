package vault

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/client"
)

var (
	testVaultIDOnce sync.Once
	testVaultID     string
)

// GetTestVaultID returns the vault ID by querying by name once and caching the result.
func GetTestVaultID(t *testing.T) string {
	testVaultIDOnce.Do(func() {
		vaultName := os.Getenv("OP_TEST_VAULT_NAME")
		if vaultName == "" {
			vaultName = "terraform-provider-acceptance-tests"
		}

		ctx := context.Background()
		client, err := client.CreateTestClient(ctx)
		if err != nil {
			t.Fatalf("failed to create test client: %v", err)
		}

		vaults, err := client.GetVaultsByTitle(ctx, vaultName)
		if err != nil {
			t.Fatalf("failed to get vault by name %q: %v", vaultName, err)
		}

		if len(vaults) == 0 {
			t.Fatalf("no vault found with name %q", vaultName)
		}
		if len(vaults) > 1 {
			t.Fatalf("multiple vaults found with name %q", vaultName)
		}

		testVaultID = vaults[0].ID
	})
	return testVaultID
}
