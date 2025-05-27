package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/testing/mock"
	sdk "github.com/1password/onepassword-sdk-go"
)

func TestClient_GetVault(t *testing.T) {
	testCases := map[string]struct {
		mockVaultAPI func() *mock.VaultAPIMock
		check        func(t *testing.T, vault *model.Vault, err error)
	}{
		"should return a vault": {
			mockVaultAPI: func() *mock.VaultAPIMock {
				m := &mock.VaultAPIMock{}
				m.On("List", context.Background()).Return([]sdk.VaultOverview{
					{
						ID:    "test-id",
						Title: "test-name",
					},
				}, nil)
				return m
			},
			check: func(t *testing.T, vault *model.Vault, err error) {
				require.NoError(t, err)
				require.Equal(t, "test-id", vault.ID)
				require.Equal(t, "test-name", vault.Title)
			},
		},
		"should return an error": {
			mockVaultAPI: func() *mock.VaultAPIMock {
				mockConnectClient := &mock.VaultAPIMock{}
				mockConnectClient.On("List", context.Background()).Return([]sdk.VaultOverview{}, errors.New("error"))
				return mockConnectClient
			},
			check: func(t *testing.T, vault *model.Vault, err error) {
				require.Error(t, err)
				require.Nil(t, vault)
			},
		},
	}

	for description, tc := range testCases {
		t.Run(description, func(t *testing.T) {
			client := &SDK{
				client: &sdk.Client{
					VaultsAPI: tc.mockVaultAPI(),
				},
			}
			vault, err := client.GetVault(context.Background(), "test-id")
			tc.check(t, vault, err)
		})
	}
}

func TestClient_GetVaultsByTitle(t *testing.T) {
	testCases := map[string]struct {
		mockVaultAPI func() *mock.VaultAPIMock
		check        func(t *testing.T, vaults []model.Vault, err error)
	}{
		"should return a single vault": {
			mockVaultAPI: func() *mock.VaultAPIMock {
				m := &mock.VaultAPIMock{}
				m.On("List", "test-name").Return([]sdk.VaultOverview{
					{
						ID:    "test-id",
						Title: "test-name",
					},
				}, nil)
				return m
			},
			check: func(t *testing.T, vaults []model.Vault, err error) {
				require.NoError(t, err)
				require.Len(t, vaults, 1)
				require.Equal(t, "test-id", vaults[0].ID)
				require.Equal(t, "test-name", vaults[0].Title)
				require.Equal(t, "test-description", vaults[0].Description)
			},
		},
		"should return a two vaults": {
			mockVaultAPI: func() *mock.VaultAPIMock {
				m := &mock.VaultAPIMock{}
				m.On("GetVaultsByTitle", "test-name").Return([]sdk.VaultOverview{
					{
						ID:    "test-id",
						Title: "test-name",
					},
					{
						ID:    "test-id-2",
						Title: "test-name-2",
					},
				}, nil)
				return m
			},
			check: func(t *testing.T, vaults []model.Vault, err error) {
				require.NoError(t, err)
				require.Len(t, vaults, 2)
				// Check the first vault
				require.Equal(t, "test-id", vaults[0].ID)
				require.Equal(t, "test-name", vaults[0].Title)
				require.Equal(t, "test-description", vaults[0].Description)
				// Check the second vault
				require.Equal(t, "test-id-2", vaults[1].ID)
				require.Equal(t, "test-name-2", vaults[1].Title)
				require.Equal(t, "test-description-2", vaults[1].Description)
			},
		},
		"should return an error": {
			mockVaultAPI: func() *mock.VaultAPIMock {
				m := &mock.VaultAPIMock{}
				m.On("GetVaultsByTitle", "test-name").Return([]sdk.VaultOverview{}, errors.New("error"))
				return m
			},
			check: func(t *testing.T, vaults []model.Vault, err error) {
				require.Error(t, err)
				require.Empty(t, vaults)
			},
		},
	}

	for description, tc := range testCases {
		t.Run(description, func(t *testing.T) {
			client := &SDK{
				client: &sdk.Client{
					VaultsAPI: tc.mockVaultAPI(),
				},
			}
			vault, err := client.GetVaultsByTitle(context.Background(), "test-name")
			tc.check(t, vault, err)
		})
	}
}
