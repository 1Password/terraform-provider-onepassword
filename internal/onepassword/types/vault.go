package types

import (
	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

func ToConnectVault(vault *sdk.VaultOverview) *connect.Vault {
	return &connect.Vault{
		ID:          vault.ID,
		Name:        vault.Title,
		Description: vault.Description,
	}
}
