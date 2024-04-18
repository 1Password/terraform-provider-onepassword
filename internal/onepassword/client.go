package onepassword

import (
	"context"

	"github.com/1Password/connect-sdk-go/onepassword"
)

// Client is a subset of connect.Client with context added.
type Client interface {
	GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error)
	GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error)
	GetItem(ctx context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error)
	GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*onepassword.Item, error)
	CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error
}

type ClientConfig struct {
	ConnectHost         string
	ConnectToken        string
	ServiceAccountToken string
	Account             string
	OpCLIPath           string
}
