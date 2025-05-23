package model

import (
	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

type Vault struct {
	ID          string
	Title       string
	Description string
}

func (v *Vault) FromConnectVault(vault *connect.Vault) {
	v.ID = vault.ID
	v.Title = vault.Name
	v.Description = vault.Description
}

func (v *Vault) FromSDKVault(vault *sdk.VaultOverview) {
	v.ID = vault.ID
	v.Title = vault.Title
	// v.Description = vault.Description // TODO: add to SDK https://gitlab.1password.io/dev/sdk/sdk-core/-/issues/435
}
