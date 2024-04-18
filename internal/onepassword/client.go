package onepassword

import (
	"context"
	"errors"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/internal/onepassword/cli"
	"github.com/1Password/terraform-provider-onepassword/internal/onepassword/connect"
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

func NewClient(config ClientConfig) (Client, error) {
	if config.ServiceAccountToken != "" || config.Account != "" {
		return cli.NewClient(config.ServiceAccountToken, config.Account, config.OpCLIPath), nil
	} else if config.ConnectHost != "" && config.ConnectToken != "" {
		return connect.NewClient(config.ConnectHost, config.ConnectToken, config.OpCLIPath), nil
	}
	return nil, errors.New("Invalid provider configuration. Either Connect credentials (\"token\" and \"url\") or Service Account (\"service_account_token\" or \"account\") credentials should be set.")
}
