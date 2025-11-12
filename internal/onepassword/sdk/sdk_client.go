package sdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	opsdk "github.com/1password/onepassword-sdk-go"
)

type Client struct {
	sdkClient *opsdk.Client
}

func (c *Client) GetVault(ctx context.Context, uuid string) (*model.Vault, error) {
	vault, err := c.sdkClient.Vaults().GetOverview(ctx, uuid)
	if err != nil {
		return nil, err
	}

	v := &model.Vault{}
	v.FromSDKVault(&vault)

	return v, nil
}

func (c *Client) GetVaultsByTitle(ctx context.Context, title string) ([]*model.Vault, error) {
	vaultList, err := c.sdkClient.Vaults().List(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.Vault
	for _, vaultOverview := range vaultList {
		fullVault, err := c.sdkClient.Vaults().GetOverview(ctx, vaultOverview.ID)
		if err != nil {
			return nil, err
		}

		if fullVault.Title == title {
			vault := &model.Vault{}
			vault.FromSDKVault(&fullVault)
			result = append(result, vault)
		}
	}

	return result, nil
}

func (c *Client) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*model.Item, error) {
	sdkItem, err := c.sdkClient.Items().Get(ctx, vaultUuid, itemUuid)
	if err != nil {
		return nil, err
	}

	item := &model.Item{}
	item.ConvertSDKItemToProviderItem(&sdkItem)

	return item, nil
}

func (c *Client) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*model.Item, error) {
	// List all items in the vault
	items, err := c.sdkClient.Items().List(ctx, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Find item matching the title
	var matchedID string
	var count int

	for _, item := range items {
		if item.Title == title {
			matchedID = item.ID
			count++
		}
	}

	// Handle no matches
	if count == 0 {
		return nil, fmt.Errorf("no item found with title %q in vault %s", title, vaultUuid)
	}

	// Handle multiple matches
	if count > 1 {
		return nil, fmt.Errorf("multiple items found with title %q in vault %s, use uuid instead", title, vaultUuid)
	}

	// Get the full item details
	return c.GetItem(ctx, matchedID, vaultUuid)
}

func (c *Client) CreateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	params := item.ConvertItemToSDKItem(vaultUuid)

	sdkItem, err := c.sdkClient.Items().Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	result := &model.Item{}
	result.ConvertSDKItemToProviderItem(&sdkItem)

	return result, nil
}

func (c *Client) UpdateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	return nil, errors.New("UpdateItem: not implemented yet")
}

func (c *Client) DeleteItem(ctx context.Context, item *model.Item, vaultUuid string) error {
	err := c.sdkClient.Items().Delete(ctx, vaultUuid, item.ID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func (c *Client) GetFileContent(ctx context.Context, file *model.ItemFile, itemUUID, vaultUUID string) ([]byte, error) {
	fileAttributes := opsdk.FileAttributes{
		Name: file.Name,
		ID:   file.ID,
		Size: uint32(file.Size),
	}

	content, err := c.sdkClient.Items().Files().Read(ctx, vaultUUID, itemUUID, fileAttributes)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return content, nil
}

func NewClient(ctx context.Context, providerUserAgent, serviceAccountToken string) (*Client, error) {
	sdkClient, err := opsdk.NewClient(ctx,
		opsdk.WithServiceAccountToken(serviceAccountToken),
		opsdk.WithIntegrationInfo("terraform-provider", "test"),
	)
	if err != nil {
		return nil, fmt.Errorf("SDK client creation failed: %w", err)
	}

	return &Client{sdkClient: sdkClient}, nil
}
