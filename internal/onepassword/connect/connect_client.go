package connect

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
)

type Config struct {
	ProviderUserAgent string
}

type Client struct {
	connectClient connect.Client
	config        Config
}

func (c *Client) GetVault(_ context.Context, uuid string) (*model.Vault, error) {
	connectVault, err := c.connectClient.GetVault(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault using connect: %w", err)
	}

	modelVault := &model.Vault{}
	modelVault.FromConnectVault(connectVault)
	return modelVault, nil
}

func (c *Client) GetVaultsByTitle(_ context.Context, title string) ([]model.Vault, error) {
	connectVaults, err := c.connectClient.GetVaultsByTitle(title)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault using connect: %w", err)
	}

	modelVaults := make([]model.Vault, len(connectVaults))
	for i, connectVault := range connectVaults {
		modelVault := model.Vault{}
		modelVault.FromConnectVault(&connectVault)
		modelVaults[i] = modelVault
	}
	return modelVaults, nil
}

// GetItem looks up an item by UUID (with retries) or by title.
// If itemUuid is a valid UUID format, it attempts to fetch the item by UUID with retries
// to handle eventual consistency issues in Connect (there can be a delay between item creation
// and when it becomes available for reading). If itemUuid is not a valid UUID format, it treats
// the parameter as a title and looks up the item by title instead.
func (c *Client) GetItem(_ context.Context, itemUuid, vaultUuid string) (*model.Item, error) {
	var connectItem *onepassword.Item

	if util.IsValidUUID(itemUuid) {
		// Try GetItemByUUID with retry for eventual consistency
		err := util.Retry404UntilCondition(context.Background(), func() (bool, error) {
			var fetchErr error
			connectItem, fetchErr = c.connectClient.GetItemByUUID(itemUuid, vaultUuid)
			if fetchErr == nil && connectItem != nil {
				return true, nil
			}
			// Return the error (404 will be retried, others returned immediately)
			return false, fetchErr
		})

		if err != nil {
			return nil, fmt.Errorf("failed to get item using connect: %w", err)
		}

		// Convert to model Item
		modelItem := &model.Item{}
		err = modelItem.FromConnectItemToModel(connectItem)
		if err != nil {
			return nil, err
		}

		return modelItem, nil
	}

	// Not a UUID, use GetItemByTitle
	connectItem, err := c.connectClient.GetItemByTitle(itemUuid, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get item using connect: %w", err)
	}

	// Convert to model Item
	modelItem := &model.Item{}
	err = modelItem.FromConnectItemToModel(connectItem)
	if err != nil {
		return nil, err
	}

	return modelItem, nil
}

func (c *Client) GetItemByTitle(_ context.Context, title string, vaultUuid string) (*model.Item, error) {
	connectItem, err := c.connectClient.GetItemByTitle(title, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get item using connect: %w", err)
	}

	// Convert to model Item
	modelItem := &model.Item{}
	err = modelItem.FromConnectItemToModel(connectItem)
	if err != nil {
		return nil, err
	}

	return modelItem, nil
}

func (c *Client) CreateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	// Convert model Item to Connect Item
	connectItem, err := item.FromModelItemToConnect()
	if err != nil {
		return nil, err
	}

	var createdItem *onepassword.Item
	err = util.RetryOnConflict(ctx, func() error {
		var createErr error
		createdItem, createErr = c.connectClient.CreateItem(connectItem, vaultUuid)
		return createErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create item using connect: %w", err)
	}
	// Wait for Connect to propagate the create to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Verify the item exists (newly created items have version 1).
	// Ignore errors from Retry404UntilCondition - if create succeeded, we return the created item even if retry times out
	_ = util.Retry404UntilCondition(ctx, func() (bool, error) {
		fetchedItem, err := c.connectClient.GetItemByUUID(createdItem.ID, vaultUuid)
		if err != nil {
			// 404 will be retried, others returned immediately
			return false, err
		}
		// Item exists, check if it has version 1 (newly created)
		if fetchedItem != nil && fetchedItem.Version == 1 {
			return true, nil
		}

		// Item exists but version doesn't match yet, continue retrying with "condition not met" error
		return false, fmt.Errorf("condition not met: item version is %d, expected 1", fetchedItem.Version)
	})

	// Convert created Connect Item back to model Item
	modelItem := &model.Item{}
	err = modelItem.FromConnectItemToModel(createdItem)
	if err != nil {
		return nil, err
	}
	return modelItem, nil
}

func (c *Client) UpdateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	// Convert model Item to Connect Item
	connectItem, err := item.FromModelItemToConnect()
	if err != nil {
		return nil, err
	}

	var updatedItem *onepassword.Item
	err = util.RetryOnConflict(ctx, func() error {
		var updateErr error
		updatedItem, updateErr = c.connectClient.UpdateItem(connectItem, vaultUuid)
		return updateErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update item using connect: %w", err)
	}

	expectedVersion := updatedItem.Version + 1 // UpdateItem doesn't return increased item version. Need to increase it manually.

	// Wait for Connect to propagate the update to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Use Retry404UntilCondition to retry until the item version matches the expected version.
	err = util.Retry404UntilCondition(ctx, func() (bool, error) {
		fetchedItem, err := c.connectClient.GetItemByUUID(updatedItem.ID, vaultUuid)
		if err != nil {
			// Return error immediately - don't retry
			return false, err
		}
		// Compare versions to verify the update has propagated
		if fetchedItem != nil && fetchedItem.Version == expectedVersion {
			return true, nil
		}
		// Version doesn't match yet, continue retrying with "condition not met" error
		return false, fmt.Errorf("condition not met: item version is %d, expected %d", fetchedItem.Version, expectedVersion)
	})
	if err != nil {
		return nil, err
	}

	// Convert updated Connect Item back to model Item
	modelItem := &model.Item{}
	err = modelItem.FromConnectItemToModel(updatedItem)
	if err != nil {
		return nil, err
	}
	return modelItem, nil
}

func (c *Client) DeleteItem(ctx context.Context, item *model.Item, vaultUuid string) error {
	// Convert model Item to Connect Item
	connectItem, err := item.FromModelItemToConnect()
	if err != nil {
		return err
	}

	err = util.RetryOnConflict(ctx, func() error {
		return c.connectClient.DeleteItem(connectItem, vaultUuid)
	})
	if err != nil {
		return fmt.Errorf("failed to delete item using connect: %w", err)
	}

	// Wait for Connect to propagate the delete to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Verify the item is deleted by checking it returns 404.
	// Ignore errors from Retry404UntilCondition - if delete succeeded, we return nil even if retry times out
	_ = util.Retry404UntilCondition(ctx, func() (bool, error) {
		_, err := c.connectClient.GetItemByUUID(item.ID, vaultUuid)
		if err != nil {
			// 404 means item is deleted, which is what we want
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				return true, nil
			}
			// Other errors are not retryable
			return false, err
		}
		// Item still exists, deletion hasn't propagated yet so retry with "condition not met" error
		return false, fmt.Errorf("condition not met: item still exists")
	})

	return nil
}

func (c *Client) GetFileContent(_ context.Context, file *model.ItemFile, itemUUID, vaultUUID string) ([]byte, error) {
	connectFile := &onepassword.File{
		ID:          file.ID,
		Name:        file.Name,
		Size:        file.Size,
		ContentPath: file.ContentPath,
	}

	// Only set Section if it exists
	// Connect expects nil if Section is not set
	if file.SectionID != "" {
		connectFile.Section = &onepassword.ItemSection{
			ID:    file.SectionID,
			Label: file.SectionLabel,
		}
	}

	content, err := c.connectClient.GetFileContent(connectFile)
	return content, err
}

func NewClient(connectHost, connectToken string, config Config) *Client {
	return &Client{
		connectClient: connect.NewClientWithUserAgent(connectHost, connectToken, config.ProviderUserAgent),
		config:        config,
	}
}
