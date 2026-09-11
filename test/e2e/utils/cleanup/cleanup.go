package cleanup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/client"
)

// RegisterItem registers a cleanup function that will delete an item even if the test fails
func RegisterItem(t *testing.T, uuidPtr *string, vaultID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if uuidPtr == nil || *uuidPtr == "" {
			return fmt.Errorf("cleanup: item UUID is empty")
		}

		itemUUID := *uuidPtr
		t.Cleanup(func() {
			// Only delete if test failed because if test succeeded Terraform already deleted it
			if !t.Failed() {
				return
			}

			ctx := context.Background()
			testClient, err := client.CreateTestClient(ctx)
			if err != nil {
				t.Logf("Cleanup: failed to create client for item %s: %v", itemUUID, err)
				return
			}

			itemToDelete := &model.Item{
				ID:      itemUUID,
				VaultID: vaultID,
			}

			// Try to delete, retry once after 1s if it fails
			err = testClient.DeleteItem(ctx, itemToDelete, vaultID)
			if err != nil {
				time.Sleep(1 * time.Second)
				err = testClient.DeleteItem(ctx, itemToDelete, vaultID)
				if err != nil {
					t.Logf("Cleanup: failed to delete item %s: %v", itemUUID, err)
					return
				}
			}
		})

		return nil
	}
}

// RegisterVault registers a cleanup function that deletes a managed vault if a test fails.
func RegisterVault(t *testing.T, uuidPtr *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if uuidPtr == nil || *uuidPtr == "" {
			return fmt.Errorf("cleanup: vault UUID is empty")
		}

		vaultUUID := *uuidPtr
		t.Cleanup(func() {
			if !t.Failed() {
				return
			}

			ctx := context.Background()
			testClient, err := client.CreateTestClient(ctx)
			if err != nil {
				t.Logf("Cleanup: failed to create client for vault %s: %v", vaultUUID, err)
				return
			}
			vaultManager, ok := testClient.(onepassword.VaultManager)
			if !ok {
				t.Logf("Cleanup: configured client cannot manage vault %s", vaultUUID)
				return
			}
			if err := vaultManager.DeleteVault(ctx, vaultUUID); err != nil {
				time.Sleep(time.Second)
				if err := vaultManager.DeleteVault(ctx, vaultUUID); err != nil {
					t.Logf("Cleanup: failed to delete vault %s: %v", vaultUUID, err)
				}
			}
		})
		return nil
	}
}
