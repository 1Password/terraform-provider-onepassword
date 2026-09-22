package model

import (
	"reflect"
	"testing"

	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

func TestFromConnectVault(t *testing.T) {
	tests := map[string]struct {
		input    *connect.Vault
		expected *Vault
	}{
		"should convert complete vault": {
			input: &connect.Vault{
				ID:          "vault1",
				Name:        "Test Vault",
				Description: "Test Description",
			},
			expected: &Vault{
				ID:          "vault1",
				Name:        "Test Vault",
				Description: "Test Description",
			},
		},
		"should handle vault with empty fields": {
			input: &connect.Vault{
				ID:          "vault1",
				Name:        "",
				Description: "",
			},
			expected: &Vault{
				ID:          "vault1",
				Name:        "",
				Description: "",
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			vault := &Vault{}
			vault.FromConnectVault(test.input)
			if !reflect.DeepEqual(vault, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, vault)
			}
		})
	}
}

func TestToConnectVault(t *testing.T) {
	tests := map[string]struct {
		input    *Vault
		expected *connect.Vault
	}{
		"should convert complete vault": {
			input: &Vault{
				ID:          "vault1",
				Name:        "Test Vault",
				Description: "Test Description",
			},
			expected: &connect.Vault{
				ID:          "vault1",
				Name:        "Test Vault",
				Description: "Test Description",
			},
		},
		"should handle vault with empty fields": {
			input: &Vault{
				ID:          "vault1",
				Name:        "",
				Description: "",
			},
			expected: &connect.Vault{
				ID:          "vault1",
				Name:        "",
				Description: "",
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := test.input.ToConnectVault()
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromSDKVault(t *testing.T) {
	tests := map[string]struct {
		input    *sdk.VaultOverview
		expected *Vault
	}{
		"should convert complete vault": {
			input: &sdk.VaultOverview{
				ID:          "vault1",
				Title:       "Test Vault",
				Description: "Test Description",
			},
			expected: &Vault{
				ID:          "vault1",
				Name:        "Test Vault",
				Description: "Test Description",
			},
		},
		"should handle vault with empty fields": {
			input: &sdk.VaultOverview{
				ID:          "vault1",
				Title:       "",
				Description: "",
			},
			expected: &Vault{
				ID:          "vault1",
				Name:        "",
				Description: "",
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			vault := &Vault{}
			vault.FromSDKVault(test.input)
			if !reflect.DeepEqual(vault, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, vault)
			}
		})
	}
}

func TestFromSDKVaultDetails(t *testing.T) {
	input := &sdk.Vault{
		ID:          "vault1",
		Title:       "Test Vault",
		Description: "Test Description",
		Access: []sdk.VaultAccess{
			{AccessorType: sdk.VaultAccessorTypeUser, AccessorUuid: "user1", Permissions: sdk.ReadItems},
			{AccessorType: sdk.VaultAccessorTypeGroup, AccessorUuid: "group1", Permissions: sdk.ReadItems | sdk.CreateItems},
		},
	}

	vault := &Vault{}
	vault.FromSDKVaultDetails(input)

	expected := &Vault{
		ID:          "vault1",
		Name:        "Test Vault",
		Description: "Test Description",
		GroupAccess: []VaultGroupAccess{{
			GroupID:     "group1",
			Permissions: sdk.ReadItems | sdk.CreateItems,
		}},
	}
	if !reflect.DeepEqual(vault, expected) {
		t.Errorf("Expected %+v, got %+v", expected, vault)
	}
}

func TestVaultSDKParams(t *testing.T) {
	vault := &Vault{Name: "Test Vault", Description: "Test Description"}

	createParams := vault.ToSDKCreateParams(false)
	if createParams.Title != vault.Name || createParams.Description == nil || *createParams.Description != vault.Description {
		t.Fatalf("Unexpected create params: %+v", createParams)
	}
	if createParams.AllowAdminsAccess == nil || *createParams.AllowAdminsAccess {
		t.Fatalf("Expected allow admins access to be false: %+v", createParams)
	}

	updateParams := vault.ToSDKUpdateParams()
	if updateParams.Title == nil || *updateParams.Title != vault.Name ||
		updateParams.Description == nil || *updateParams.Description != vault.Description {
		t.Fatalf("Unexpected update params: %+v", updateParams)
	}
}
