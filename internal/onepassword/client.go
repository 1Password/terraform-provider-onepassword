package onepassword

import (
	"context"
	"errors"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/connect"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

// Client is a subset of connect.Client with context added.
type Client interface {
	GetVault(ctx context.Context, uuid string) (*model.Vault, error)
	GetVaultsByTitle(ctx context.Context, title string) ([]*model.Vault, error)
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
	OpCLIPath           string
	ProviderUserAgent   string
}

func NewClient(config ClientConfig) (Client, error) {
	// if config.ServiceAccountToken != "" || config.Account != "" {
	// 	return cli.NewClient(config.ServiceAccountToken, config.Account, config.OpCLIPath), nil
	// } else
	if config.ConnectHost != "" && config.ConnectToken != "" {
		return connect.NewClient(config.ConnectHost, config.ConnectToken, connect.Config{
			ProviderUserAgent: config.ProviderUserAgent,
		}), nil
	}
	return nil, errors.New("Invalid provider configuration. Either Connect credentials (\"token\" and \"url\") or Service Account (\"service_account_token\" or \"account\") credentials should be set.")
}
