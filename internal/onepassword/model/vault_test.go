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
