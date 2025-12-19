package onepassword

import (
	"context"
	"errors"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/connect"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/sdk"
)

// Client is a subset of connect.Client with context added.
type Client interface {
	GetVault(ctx context.Context, uuid string) (*model.Vault, error)
	GetVaultsByTitle(ctx context.Context, title string) ([]model.Vault, error)
	GetItem(ctx context.Context, itemUuid, vaultUuid string) (*model.Item, error)
	GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*model.Item, error)
	CreateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error)
	UpdateItem(ctx context.Context, item *model.Item, vaultUuid string) (*model.Item, error)
	DeleteItem(ctx context.Context, item *model.Item, vaultUuid string) error
	GetFileContent(ctx context.Context, file *model.ItemFile, itemUUid, vaultUuid string) ([]byte, error)
}

type ClientConfig struct {
	ConnectHost         string
	ConnectToken        string
	ServiceAccountToken string
	Account             string
	ProviderUserAgent   string
	MaxRetries          int
}

func NewClient(ctx context.Context, config ClientConfig) (Client, error) {
	if config.ServiceAccountToken != "" || config.Account != "" {
		return sdk.NewClient(ctx, sdk.SDKConfig{
			ProviderUserAgent:   config.ProviderUserAgent,
			ServiceAccountToken: config.ServiceAccountToken,
			Account:             config.Account,
			MaxRetries:          config.MaxRetries,
		})
	} else if config.ConnectHost != "" && config.ConnectToken != "" {
		return connect.NewClient(config.ConnectHost, config.ConnectToken, connect.Config{
			ProviderUserAgent: config.ProviderUserAgent,
			MaxRetries:        config.MaxRetries,
		}), nil
	}
	return nil, errors.New("Invalid provider configuration. Either Connect credentials (\"connect_token\" and \"connect_url\") or Service Account (\"service_account_token\") or \"account\"  should be set.")
}
