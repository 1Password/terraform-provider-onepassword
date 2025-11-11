package sdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/types"
	opsdk "github.com/1password/onepassword-sdk-go"
)

type Client struct {
	sdkClient *opsdk.Client
}

func (c *Client) GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error) {
	vault, err := c.sdkClient.Vaults().GetOverview(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return types.ToConnectVault(&vault), nil
}

func (c *Client) GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error) {
	vaults, err := c.sdkClient.Vaults().List(ctx)
	if err != nil {
		return nil, err
	}

	var result []onepassword.Vault
	for _, v := range vaults {
		if v.Title == title {
			result = append(result, *types.ToConnectVault(&v))
		}
	}

	return result, nil
}

func (c *Client) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error) {
	sdkItem, err := c.sdkClient.Items().Get(ctx, vaultUuid, itemUuid)
	if err != nil {
		return nil, err
	}

	return types.ToConnectItem(&sdkItem), nil
}

func (c *Client) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*onepassword.Item, error) {
	return nil, errors.New("GetItemByTitle: not implemented yet")
}

func (c *Client) CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return nil, errors.New("CreateItem: not implemented yet")
}

func (c *Client) UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return nil, errors.New("UpdateItem: not implemented yet")
}

func (c *Client) DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error {
	return errors.New("DeleteItem: not implemented yet")
}

func (c *Client) GetFileContent(ctx context.Context, file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
	return nil, errors.New("GetFileContent: not implemented yet")
}

func NewClient(ctx context.Context, providerUserAgent string) (*Client, error) {
	sdkClient, err := opsdk.NewClient(ctx,
		opsdk.WithDesktopAppIntegration(""),
		opsdk.WithIntegrationInfo("terraform-provider", providerUserAgent),
	)

	if err != nil {
		return nil, fmt.Errorf("SDK client creation failed: %w", err)
	}

	return &Client{sdkClient: sdkClient}, nil
}
