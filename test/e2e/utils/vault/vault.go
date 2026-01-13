package vault

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/client"
)

var (
	testVaultIDOnce sync.Once
	testVaultID     string
	testVaultIDErr  error
)

const defaultTestVaultID = "bbucuyq2nn4fozygwttxwizpcy"

// InitTestVaultID initializes the vault ID. Returns error if initialization fails.
func InitTestVaultID() error {
	testVaultIDOnce.Do(func() {
		vaultName := os.Getenv("OP_TEST_VAULT_NAME")
		if vaultName == "" {
			testVaultID = defaultTestVaultID
			return
		}

		ctx := context.Background()
		client, err := client.CreateTestClient(ctx)
		if err != nil {
			testVaultIDErr = fmt.Errorf("failed to create test client: %w", err)
			return
		}

		vaults, err := client.GetVaultsByTitle(ctx, vaultName)
		if err != nil {
			testVaultIDErr = fmt.Errorf("failed to get vault by name %q: %w", vaultName, err)
			return
		}

		if len(vaults) == 0 {
			testVaultIDErr = fmt.Errorf("no vault found with name %q", vaultName)
			return
		}
		if len(vaults) > 1 {
			testVaultIDErr = fmt.Errorf("multiple vaults found with name %q", vaultName)
			return
		}

		testVaultID = vaults[0].ID
	})

	return testVaultIDErr
}

// GetTestVaultID returns the vault ID.
func GetTestVaultID(t *testing.T) string {
	err := InitTestVaultID()
	if err != nil {
		t.Fatalf("failed to get test vault ID: %v", err)
	}

	return testVaultID
}
