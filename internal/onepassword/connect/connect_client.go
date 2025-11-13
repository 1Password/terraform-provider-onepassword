package connect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"

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

func (c *Client) GetVault(_ context.Context, uuid string) (*onepassword.Vault, error) {
	return c.connectClient.GetVault(uuid)
}

func (c *Client) GetVaultsByTitle(_ context.Context, title string) ([]onepassword.Vault, error) {
	return c.connectClient.GetVaultsByTitle(title)
}

// GetItem looks up an item by UUID (with retries) or by title.
// If itemUuid is a valid UUID format, it attempts to fetch the item by UUID with retries
// to handle eventual consistency issues in Connect (there can be a delay between item creation
// and when it becomes available for reading). If itemUuid is not a valid UUID format, it treats
// the parameter as a title and looks up the item by title instead.
func (c *Client) GetItem(_ context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error) {
	if util.IsValidUUID(itemUuid) {
		// Try GetItemByUUID with retry for eventual consistency
		var item *onepassword.Item
		var err error
		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			}
			item, err = c.connectClient.GetItemByUUID(itemUuid, vaultUuid)
			if item != nil {
				return item, nil // item is found by UUID, return
			}
			// If error is not 404, don't retry
			if err != nil && !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
				return nil, err
			}
		}
		return nil, err
	}

	// Not a UUID, use GetItemByTitle
	return c.connectClient.GetItemByTitle(itemUuid, vaultUuid)
}

func (c *Client) GetItemByTitle(_ context.Context, title string, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.GetItemByTitle(title, vaultUuid)
}

func (c *Client) CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	createdItem, err := c.connectClient.CreateItem(item, vaultUuid)
	if err != nil {
		return nil, err
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

	return createdItem, nil
}

func (c *Client) UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	updatedItem, err := c.connectClient.UpdateItem(item, vaultUuid)
	if err != nil {
		return nil, err
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

	return updatedItem, nil
}

func (c *Client) DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error {
	err := c.connectClient.DeleteItem(item, vaultUuid)
	if err != nil {
		return err
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

func (w *Client) GetFileContent(_ context.Context, file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
	return w.connectClient.GetFileContent(file)
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
