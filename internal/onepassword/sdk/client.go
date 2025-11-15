package sdk

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/1password/onepassword-sdk-go"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

type Client struct {
	sdkClient *sdk.Client
}

type SDKConfig struct {
	ProviderUserAgent   string
	ServiceAccountToken string
	Account             string
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

	modelItem := &model.Item{}
	modelItem.FromSDKItemToModel(&sdkItem)
	return modelItem, nil
}

func (c *Client) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*model.Item, error) {
	items, err := c.sdkClient.Items().List(ctx, vaultUuid)
	if err != nil {
		return nil, err
	}

	var matchedID string
	var count int

	for _, item := range items {
		if item.Title == title {
			matchedID = item.ID
			count++
		}
	}

	return c.GetItem(ctx, matchedID, vaultUuid)
}

func (c *Client) CreateItem(ctx context.Context, item *model.Item) (*model.Item, error) {
	params := item.FromModelItemToSDKCreateParams()

	sdkItem, err := c.sdkClient.Items().Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	modelItem := &model.Item{}
	modelItem.FromSDKItemToModel(&sdkItem)
	return modelItem, nil
}

func (c *Client) UpdateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	currentItem, err := c.sdkClient.Items().Get(ctx, vaultUuid, item.ID)
	if err != nil {
		return nil, err
	}

	params := item.FromModelItemToSDKCreateParams()
	currentItem.Title = params.Title
	currentItem.Category = params.Category
	currentItem.Fields = params.Fields
	currentItem.Sections = params.Sections
	currentItem.Tags = params.Tags
	currentItem.Websites = params.Websites
	if params.Notes != nil {
		currentItem.Notes = *params.Notes
	}

	updatedItem, err := c.sdkClient.Items().Put(ctx, currentItem)
	if err != nil {
		return nil, err
	}

	// Convert back to provider model
	modelItem := &model.Item{}
	modelItem.FromSDKItemToModel(&updatedItem)
	return modelItem, nil
}

func (c *Client) DeleteItem(ctx context.Context, item *model.Item, vaultUuid string) error {
	err := c.sdkClient.Items().Delete(ctx, vaultUuid, item.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFileContent(ctx context.Context, file *model.ItemFile, itemUUID, vaultUUID string) ([]byte, error) {
	fileAttributes := sdk.FileAttributes{
		Name: file.Name,
		ID:   file.ID,
		Size: uint32(file.Size),
	}

	content, err := c.sdkClient.Items().Files().Read(ctx, vaultUUID, itemUUID, fileAttributes)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func NewClient(ctx context.Context, config SDKConfig) (*Client, error) {
	var sdkClient *sdk.Client
	var err error

	integrationName, integrationVersion, _ := strings.Cut(config.ProviderUserAgent, "/")

	// Initialize with service account token if provided, otherwise use desktop integration
	if config.ServiceAccountToken != "" {
		sdkClient, err = sdk.NewClient(ctx,
			sdk.WithServiceAccountToken(config.ServiceAccountToken),
			sdk.WithIntegrationInfo(integrationName, integrationVersion),
		)
		if err != nil {
			return nil, fmt.Errorf("SDK client creation with service account failed: %w", err)
		}
	} else {
		// Fall back to desktop integration
		sdkClient, err = sdk.NewClient(ctx,
			sdk.WithDesktopAppIntegration(config.Account),
			sdk.WithIntegrationInfo(integrationName, integrationVersion),
		)
		if err != nil {
			return nil, fmt.Errorf("SDK client creation with desktop integration failed: %w", err)
		}
	}

	return &Client{sdkClient: sdkClient}, nil
}
