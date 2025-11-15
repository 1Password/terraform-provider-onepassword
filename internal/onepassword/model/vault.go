package model

import (
	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

type Vault struct {
	ID          string
	Name        string
	Description string
}

func (v *Vault) FromConnectVault(vault *connect.Vault) {
	v.ID = vault.ID
	v.Name = vault.Name
	v.Description = vault.Description
}

func (v *Vault) FromSDKVault(vault *sdk.VaultOverview) {
	v.ID = vault.ID
	v.Name = vault.Title
	v.Description = vault.Description
}
