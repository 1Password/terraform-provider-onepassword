package connect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
)

type Config struct {
	// MaxRetries is the maximum number of retry attempts when waiting for Connect to
	// propagate changes. The wait function uses exponential backoff between retries.
	MaxRetries        int
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
	var err error

	if util.IsValidUUID(itemUuid) {
		// Try GetItemByUUID with retry for eventual consistency
		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			}
			connectItem, err = c.connectClient.GetItemByUUID(itemUuid, vaultUuid)

			if err == nil && connectItem != nil {
				// Convert to model Item
				modelItem := &model.Item{}
				err := modelItem.FromConnectItemToModel(connectItem)
				if err != nil {
					return nil, err
				}

				return modelItem, nil
			}
			// If error is not 404, don't retry
			if err != nil && !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("failed to get item using connect: %w", err)
			}
		}
		return nil, err
	}

	// Not a UUID, use GetItemByTitle
	connectItem, err = c.connectClient.GetItemByTitle(itemUuid, vaultUuid)
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

	createdItem, err := c.connectClient.CreateItem(connectItem, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to create item using connect: %w", err)
	}

	// Wait for Connect to propagate the create to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Verify the item exists (newly created items have version 1).
	// Ignore errors from wait - if create succeeded, we return the created item even if wait times out
	_ = c.wait(ctx, createdItem.ID, vaultUuid, func(fetchedItem *onepassword.Item, err error) (bool, error) {
		if err != nil {
			// If error is 404, item not available yet, continue retrying
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				return false, nil
			}
			// Other errors are not retryable
			return false, err
		}
		// Item exists, check if it has version 1 (newly created)
		if fetchedItem != nil && fetchedItem.Version == 1 {
			return true, nil
		}
		// Item exists but version doesn't match yet, continue retrying
		return false, nil
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

	updatedItem, err := c.connectClient.UpdateItem(connectItem, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to update item using connect: %w", err)
	}

	expectedVersion := updatedItem.Version + 1 // UpdateItem doesn't return increased item version. Need to increase it manually.

	// Wait for Connect to propagate the update to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Verify the item version matches the expected version.
	err = c.wait(ctx, updatedItem.ID, vaultUuid, func(fetchedItem *onepassword.Item, err error) (bool, error) {
		if err != nil {
			// For updates, any error (including 404) means something is wrong since the item should exist
			// Return error immediately - don't retry
			return false, err
		}
		// Compare versions to verify the update has propagated
		if fetchedItem != nil && fetchedItem.Version == expectedVersion {
			return true, nil
		}
		// Version doesn't match yet, continue retrying
		return false, nil
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

	err = c.connectClient.DeleteItem(connectItem, vaultUuid)
	if err != nil {
		return fmt.Errorf("failed to delete item using connect: %w", err)
	}

	// Wait for Connect to propagate the delete to the local SQLite database.
	// The sync service needs time to sync changes from the remote service to the local database.
	// Verify the item is deleted by checking it returns 404.
	// Ignore errors from wait - if delete succeeded, we return nil even if wait times out
	_ = c.wait(ctx, item.ID, vaultUuid, func(fetchedItem *onepassword.Item, err error) (bool, error) {
		if err != nil {
			// 404 means item is deleted, which is what we want
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				return true, nil
			}
			// Other errors are not retryable
			return false, err
		}
		// Item still exists, deletion hasn't propagated yet
		return false, nil
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

// waitCondition is a function that checks if a condition is met.
// It returns (done bool, err error) where:
//   - done=true means the condition is met and we can stop waiting
//   - done=false means we should continue retrying
//   - err!=nil means a non-retryable error occurred
type waitCondition func(fetchedItem *onepassword.Item, err error) (bool, error)

// wait waits for a condition to be met by polling the item with exponential backoff.
// The condition function is called with the fetched item (or nil) and any error from the fetch.
// This ensures the sync service has propagated changes from the remote service to the local database.
// Returns an error if the condition function returns a non-retryable error or if max retry attempts are reached.
func (c *Client) wait(ctx context.Context, itemUUID, vaultUUID string, condition waitCondition) error {
	maxAttempts := c.config.MaxRetries

	for attempt := 0; attempt < maxAttempts; attempt++ {
		fetchedItem, err := c.connectClient.GetItemByUUID(itemUUID, vaultUUID)

		// Check the condition
		done, conditionErr := condition(fetchedItem, err)
		if conditionErr != nil {
			// Non-retryable error, return it
			return conditionErr
		}
		if done {
			// Condition met, stop waiting
			return nil
		}

		// Condition not met yet, continue retrying
		// Exponential backoff: 50ms, 100ms, 200ms, 400ms, etc., capped at 500ms
		backoff := time.Duration(50*(1<<uint(attempt))) * time.Millisecond
		if backoff > 500*time.Millisecond {
			backoff = 500 * time.Millisecond
		}

		// Don't sleep on the last attempt
		if attempt < maxAttempts-1 {
			time.Sleep(backoff)
		}
	}
	return fmt.Errorf("max retry attempts (%d) reached waiting for Connect sync service to propagate changes to local database", maxAttempts)
}

func NewClient(connectHost, connectToken string, config Config) *Client {
	// Set the default max retries to 10 if not provided
	if config.MaxRetries == 0 {
		config.MaxRetries = 10
	}

	return &Client{
		connectClient: connect.NewClientWithUserAgent(connectHost, connectToken, config.ProviderUserAgent),
		config:        config,
	}
}
