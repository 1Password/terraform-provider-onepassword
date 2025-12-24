package cleanup

import (
	"context"
	"testing"
	"time"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/client"
)

// RegisterItemCleanup registers a cleanup function that will delete an item even if the test fails.
func RegisterItemCleanup(t *testing.T, itemUUID, vaultID string) {
	if itemUUID == "" {
		return
	}

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
			_ = testClient.DeleteItem(ctx, itemToDelete, vaultID)
		}
	})
}
