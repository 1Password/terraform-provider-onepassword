package sdk

import (
	"context"

	op "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

type Config struct {
	ServiceAccountToken string
	IntegrationName     string
	IntegrationVersion  string
}

type SDK struct {
	client *sdk.Client
}

func NewClient(config Config) (*SDK, error) {
	// Authenticates with your service account token and connects to 1Password.
	client, err := sdk.NewClient(context.Background(),
		sdk.WithServiceAccountToken(config.ServiceAccountToken),
		sdk.WithIntegrationInfo(config.IntegrationName, config.IntegrationVersion),
	)
	if err != nil {
		return nil, err
	}

	return &SDK{
		client: client,
	}, nil
}

func (sdk *SDK) GetVault(ctx context.Context, uuid string) (*op.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) GetVaultsByTitle(ctx context.Context, title string) ([]op.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) CreateItem(ctx context.Context, item *op.Item, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) UpdateItem(ctx context.Context, item *op.Item, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) DeleteItem(ctx context.Context, item *op.Item, vaultUuid string) error {
	//TODO implement me
	panic("implement me")
}

func (sdk *SDK) GetFileContent(ctx context.Context, file *op.File, itemUUid, vaultUuid string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
