package connect

import (
	"context"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
)

type Client struct {
	connectClient connect.Client
}

func (c *Client) GetVault(_ context.Context, uuid string) (*onepassword.Vault, error) {
	return c.connectClient.GetVault(uuid)
}

func (c *Client) GetVaultsByTitle(_ context.Context, title string) ([]onepassword.Vault, error) {
	return c.connectClient.GetVaultsByTitle(title)
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

func NewClient(connectHost, connectToken, providerUserAgent string) *Client {
	return &Client{connectClient: connect.NewClientWithUserAgent(connectHost, connectToken, providerUserAgent)}
}
