package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

func TestFromConnectVault(t *testing.T) {
	connectVault := &connect.Vault{
		ID:          "test-id",
		Name:        "test-name",
		Description: "test-description",
	}

	vault := &Vault{}
	vault.FromConnectVault(connectVault)

	require.Equal(t, connectVault.ID, vault.ID)
	require.Equal(t, connectVault.Name, vault.Title)
	require.Equal(t, connectVault.Description, vault.Description)
}

func TestFromSDKVault(t *testing.T) {
	sdkVault := &sdk.VaultOverview{
		ID:    "test-id",
		Title: "test-name",
		// Description: "test-description", // TODO: uncomment when added to SDK
	}

	vault := &Vault{}
	vault.FromSDKVault(sdkVault)

	require.Equal(t, sdkVault.ID, vault.ID)
	require.Equal(t, sdkVault.Title, vault.Title)
	//require.Equal(t, sdkVault.Description, vault.Description) // TODO: uncomment when added to SDK
}
