package onepassword

import (
	"context"
	"errors"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/connect"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/sdk"
)

// Client is a subset of connect.Client with context added.
type Client interface {
	GetVault(ctx context.Context, uuid string) (*model.Vault, error)
	GetVaultsByTitle(ctx context.Context, title string) ([]model.Vault, error)
	GetItem(ctx context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error)
	GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*onepassword.Item, error)
	CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error
	GetFileContent(ctx context.Context, file *onepassword.File, itemUUid, vaultUuid string) ([]byte, error)
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
		return sdk.NewClient(sdk.Config{
			ServiceAccountToken: config.ServiceAccountToken,
		})
	} else if config.ConnectHost != "" && config.ConnectToken != "" {
		return connect.NewClient(config.ConnectHost, config.ConnectToken, "TFP"), nil
	}
	return nil, errors.New("Invalid provider configuration. Either Connect credentials (\"token\" and \"url\") or Service Account (\"service_account_token\" or \"account\") credentials should be set.")
}
