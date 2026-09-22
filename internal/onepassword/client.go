package onepassword

import (
	"context"
	"errors"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/connect"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/sdk"
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
	// GetItems reads several items from a vault at once. Each element of itemTitlesOrIDs may be
	// an item title or an item ID. It returns exactly len(itemTitlesOrIDs) items, one per input
	// and in input order, so items[i] always corresponds to itemTitlesOrIDs[i]. The same item is
	// returned more than once if it is requested under more than one identifier. Any item that
	// cannot be read fails the whole call; a partial slice is never returned.
	GetItems(ctx context.Context, vaultUUID string, itemTitlesOrIDs []string) ([]*model.Item, error)
	GetFileContent(ctx context.Context, file *model.ItemFile, itemUUid, vaultUuid string) ([]byte, error)
	// GetEnvironmentVariables reads variables from a 1Password Environment. Only supported when using the 1Password SDK (service account or desktop app); not supported with 1Password Connect.
	GetEnvironmentVariables(ctx context.Context, environmentID string) ([]model.EnvironmentVariable, error)
}

type ClientConfig struct {
	ConnectHost         string
	ConnectToken        string
	ServiceAccountToken string
	Account             string
	ProviderUserAgent   string
}

func NewClient(ctx context.Context, config ClientConfig) (Client, error) {
	if config.ServiceAccountToken != "" || config.Account != "" {
		return sdk.NewClient(ctx, sdk.SDKConfig{
			ProviderUserAgent:   config.ProviderUserAgent,
			ServiceAccountToken: config.ServiceAccountToken,
			Account:             config.Account,
		})
	} else if config.ConnectHost != "" && config.ConnectToken != "" {
		return connect.NewClient(config.ConnectHost, config.ConnectToken, connect.Config{
			ProviderUserAgent: config.ProviderUserAgent,
		}), nil
	}
	return nil, errors.New("Invalid provider configuration. Either Connect credentials (\"connect_token\" and \"connect_url\") or Service Account (\"service_account_token\") or \"account\"  should be set.")
}
