package sdk

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/1password/onepassword-sdk-go"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
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
		return nil, fmt.Errorf("failed to get vault using sdk: %w", err)
	}

	v := &model.Vault{}
	v.FromSDKVault(&vault)

	return v, nil
}

func (c *Client) GetVaultsByTitle(ctx context.Context, title string) ([]model.Vault, error) {
	decryptDetails := true
	vaultList, err := c.sdkClient.Vaults().List(ctx, sdk.VaultListParams{DecryptDetails: &decryptDetails})
	if err != nil {
		return nil, fmt.Errorf("failed to get vaults using sdk: %w", err)
	}

	var result []model.Vault
	for _, vault := range vaultList {
		if vault.Title == title {
			var modelVault model.Vault
			modelVault.FromSDKVault(&vault)
			result = append(result, modelVault)
		}
	}

	return result, nil
}

// GetItem looks up an item by UUID or by title.
// If itemUuid is a valid UUID format, it attempts to fetch the item by UUID.
// If itemUuid is not a valid UUID format, it treats the parameter as a title
// and looks up the item by title instead.
func (c *Client) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*model.Item, error) {
	if util.IsValidUUID(itemUuid) {
		// Valid UUID, use GetItem directly
		sdkItem, err := c.sdkClient.Items().Get(ctx, vaultUuid, itemUuid)
		if err != nil {
			return nil, fmt.Errorf("failed to get item using sdk: %w", err)
		}

		modelItem := &model.Item{}
		err = modelItem.FromSDKItemToModel(&sdkItem)
		if err != nil {
			return nil, fmt.Errorf("sdk.GetItem failed to convert item using sdk: %w", err)
		}
		return modelItem, nil
	}

	// Not a UUID, use GetItemByTitle
	return c.GetItemByTitle(ctx, itemUuid, vaultUuid)
}

func (c *Client) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*model.Item, error) {
	items, err := c.sdkClient.Items().List(ctx, vaultUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get item using sdk: %w", err)
	}

	var matchedID string
	var count int

	for _, item := range items {
		if item.Title == title {
			matchedID = item.ID
			count++
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("found %d item(s) in vault %q with title %q", count, vaultUuid, title)
	}

	sdkItem, err := c.sdkClient.Items().Get(ctx, vaultUuid, matchedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item using sdk: %w", err)
	}

	modelItem := &model.Item{}
	err = modelItem.FromSDKItemToModel(&sdkItem)
	if err != nil {
		return nil, fmt.Errorf("sdk.GetItemByTitle failed to convert item using sdk: %w", err)
	}

	return modelItem, nil
}

func (c *Client) CreateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error) {
	params := item.FromModelItemToSDKCreateParams()

	if params.VaultID != vaultUuid {
		return nil, fmt.Errorf("vault UUID mismatch: item has %s but %s was provided", params.VaultID, vaultUuid)
	}

	var sdkItem sdk.Item
	err := util.RetryOnConflict(ctx, func() error {
		var createErr error
		sdkItem, createErr = c.sdkClient.Items().Create(ctx, params)
		return createErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create item using sdk: %w", err)
	}

	modelItem := &model.Item{}
	err = modelItem.FromSDKItemToModel(&sdkItem)
	if err != nil {
		return nil, fmt.Errorf("sdk.CreateItem failed to convert item using sdk: %w", err)
	}
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

	var updatedItem sdk.Item
	err = util.RetryOnConflict(ctx, func() error {
		var updateErr error
		updatedItem, updateErr = c.sdkClient.Items().Put(ctx, currentItem)
		return updateErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update item using sdk: %w", err)
	}

	// Convert back to provider model
	modelItem := &model.Item{}
	err = modelItem.FromSDKItemToModel(&updatedItem)
	if err != nil {
		return nil, fmt.Errorf("sdk.UpdateItem failed to convert item using sdk: %w", err)
	}
	return modelItem, nil
}

func (c *Client) DeleteItem(ctx context.Context, item *model.Item, vaultUuid string) error {
	err := util.RetryOnConflict(ctx, func() error {
		return c.sdkClient.Items().Delete(ctx, vaultUuid, item.ID)
	})
	if err != nil {
		return fmt.Errorf("failed to delete item using sdk: %w", err)
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
		return nil, fmt.Errorf("failed to read file using sdk: %w", err)
	}

	return content, nil
}

func NewClient(ctx context.Context, config SDKConfig) (*Client, error) {
	var sdkClient *sdk.Client
	var err error

	integrationName, integrationVersion, found := strings.Cut(config.ProviderUserAgent, "/")
	if !found {
		return nil, fmt.Errorf("invalid ProviderUserAgent format: expected 'name/version', got %q", config.ProviderUserAgent)
	}

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
