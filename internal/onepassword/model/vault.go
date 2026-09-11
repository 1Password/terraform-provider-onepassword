package model

import (
	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

const (
	VaultPermissionNoAccess           = sdk.NoAccess
	VaultPermissionManageVault        = sdk.ManageVault
	VaultPermissionRevealItemPassword = sdk.RevealItemPassword
	VaultPermissionReadItems          = sdk.ReadItems
	VaultPermissionUpdateItems        = sdk.UpdateItems
	VaultPermissionCreateItems        = sdk.CreateItems
	VaultPermissionArchiveItems       = sdk.ArchiveItems
	VaultPermissionDeleteItems        = sdk.DeleteItems
	VaultPermissionUpdateItemHistory  = sdk.UpdateItemHistory
	VaultPermissionSendItems          = sdk.SendItems
	VaultPermissionImportItems        = sdk.ImportItems
	VaultPermissionExportItems        = sdk.ExportItems
	VaultPermissionPrintItems         = sdk.PrintItems
)

type Vault struct {
	ID          string
	Name        string
	Description string
	GroupAccess []VaultGroupAccess
}

type VaultGroupAccess struct {
	GroupID     string
	Permissions uint32
}

func (v *Vault) FromConnectVault(vault *connect.Vault) {
	v.ID = vault.ID
	v.Name = vault.Name
	v.Description = vault.Description
}

func (v *Vault) ToConnectVault() *connect.Vault {
	return &connect.Vault{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
	}
}

func (v *Vault) FromSDKVault(vault *sdk.VaultOverview) {
	v.ID = vault.ID
	v.Name = vault.Title
	v.Description = vault.Description
}

func (v *Vault) FromSDKVaultDetails(vault *sdk.Vault) {
	v.ID = vault.ID
	v.Name = vault.Title
	v.Description = vault.Description
	v.GroupAccess = nil

	for _, access := range vault.Access {
		if access.AccessorType != sdk.VaultAccessorTypeGroup {
			continue
		}
		v.GroupAccess = append(v.GroupAccess, VaultGroupAccess{
			GroupID:     access.AccessorUuid,
			Permissions: access.Permissions,
		})
	}
}

func (v *Vault) ToSDKCreateParams(allowAdminsAccess bool) sdk.VaultCreateParams {
	return sdk.VaultCreateParams{
		Title:             v.Name,
		Description:       &v.Description,
		AllowAdminsAccess: &allowAdminsAccess,
	}
}

func (v *Vault) ToSDKUpdateParams() sdk.VaultUpdateParams {
	return sdk.VaultUpdateParams{
		Title:       &v.Name,
		Description: &v.Description,
	}
}
