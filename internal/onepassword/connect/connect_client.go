package connect

import (
	"context"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
)

type Client struct {
	connectClient connect.Client
}

func (c *Client) GetVault(_ context.Context, uuid string) (*model.Vault, error) {
	connectVault, err := c.connectClient.GetVault(uuid)
	if err != nil {
		return nil, err
	}
	var vault model.Vault
	vault.FromConnectVault(connectVault)

	return &vault, nil
}

func (c *Client) GetVaultsByTitle(_ context.Context, title string) ([]model.Vault, error) {
	connectVaults, err := c.connectClient.GetVaultsByTitle(title)
	if err != nil {
		return nil, err
	}

	var vaults []model.Vault
	for _, connectVault := range connectVaults {
		var vault model.Vault
		vault.FromConnectVault(&connectVault)
		vaults = append(vaults, vault)
	}

	return vaults, nil
}

func (c *Client) GetItem(_ context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.GetItem(itemUuid, vaultUuid)
}

func (c *Client) GetItemByTitle(_ context.Context, title string, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.GetItemByTitle(title, vaultUuid)
}

func (c *Client) CreateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.CreateItem(item, vaultUuid)
}

func (c *Client) UpdateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return c.connectClient.UpdateItem(item, vaultUuid)
}

func (c *Client) DeleteItem(_ context.Context, item *onepassword.Item, vaultUuid string) error {
	return c.connectClient.DeleteItem(item, vaultUuid)
}

func (c *Client) GetFileContent(_ context.Context, file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
	return c.connectClient.GetFileContent(file)
}

func NewClient(connectHost, connectToken, providerUserAgent string) *Client {
	return &Client{connectClient: connect.NewClientWithUserAgent(connectHost, connectToken, providerUserAgent)}
}
