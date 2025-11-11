package connect

import (
	"context"
	"strings"
	"time"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
)

type Config struct {
	EventualConsistencyDelay time.Duration
	ProviderUserAgent        string
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

func (c *Client) CreateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.CreateItem(item, vaultUuid)
}

func (c *Client) UpdateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	updatedItem, err := c.connectClient.UpdateItem(item, vaultUuid)
	if err != nil {
		return nil, err
	}

	// Wait for Connect to propagate the update before Terraform's automatic Read runs.
	// Connect has eventual consistency, so we need to wait to ensure the subsequent Read
	// gets the updated values and prevents refresh plan issues.
	time.Sleep(c.config.EventualConsistencyDelay)

	return updatedItem, nil
}

func (c *Client) DeleteItem(_ context.Context, item *onepassword.Item, vaultUuid string) error {
	err := c.connectClient.DeleteItem(item, vaultUuid)
	if err != nil {
		return err
	}

	// Wait for Connect to propagate the delete to ensure eventual consistency.
	// This helps prevent race conditions if the same item is recreated immediately
	// or if there are parallel operations.
	time.Sleep(c.config.EventualConsistencyDelay)

	return nil
}

func (w *Client) GetFileContent(_ context.Context, file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
	return w.connectClient.GetFileContent(file)
}

func NewClient(connectHost, connectToken string, config Config) *Client {
	// Set the default eventual consistency delay to 500ms if not provided
	if config.EventualConsistencyDelay == 0 {
		config.EventualConsistencyDelay = 500 * time.Millisecond
	}

	return &Client{
		connectClient: connect.NewClientWithUserAgent(connectHost, connectToken, config.ProviderUserAgent),
		config:        config,
	}
}
