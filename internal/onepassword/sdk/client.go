package sdk

import (
	"context"
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
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

func (s *SDK) GetVault(ctx context.Context, uuid string) (*model.Vault, error) {
	vaults, err := s.client.VaultsAPI.List(ctx)
	if err != nil {
		return nil, err
	}

	var targetVault *sdk.VaultOverview
	for _, vault := range vaults {
		if vault.ID == uuid {
			targetVault = &vault
			break
		}
	}

	if targetVault == nil {
		return nil, fmt.Errorf("vault with uuid %s not found", uuid)
	}

	var vault model.Vault
	vault.FromSDKVault(targetVault)

	return &vault, nil
}

func (s *SDK) GetVaultsByTitle(ctx context.Context, title string) ([]model.Vault, error) {
	sdkVaults, err := s.client.VaultsAPI.List(ctx)
	if err != nil {
		return nil, err
	}

	// find the vaults with the provided title
	var sdkVaultsWithSameTitle []sdk.VaultOverview
	for _, sdkVault := range sdkVaults {
		if sdkVault.Title == title {
			sdkVaultsWithSameTitle = append(sdkVaultsWithSameTitle, sdkVault)
		}
	}

	// map the sdk vaults to model vaults
	var vaults []model.Vault
	for _, sdkVault := range sdkVaultsWithSameTitle {
		var vault model.Vault
		vault.FromSDKVault(&sdkVault)
		vaults = append(vaults, vault)
	}

	return vaults, nil
}

func (s *SDK) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SDK) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SDK) CreateItem(ctx context.Context, item *op.Item, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SDK) UpdateItem(ctx context.Context, item *op.Item, vaultUuid string) (*op.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SDK) DeleteItem(ctx context.Context, item *op.Item, vaultUuid string) error {
	//TODO implement me
	panic("implement me")
}

func (s *SDK) GetFileContent(ctx context.Context, file *op.File, itemUUid, vaultUuid string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
