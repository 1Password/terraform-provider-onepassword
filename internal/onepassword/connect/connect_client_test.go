package connect

import (
	"context"
	"errors"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/testing/mock"
)

func TestClient_GetVault(t *testing.T) {
	testCases := map[string]struct {
		mockClient func() *mock.ConnectClientMock
		check      func(t *testing.T, vault *model.Vault, err error)
	}{
		"should return a vault": {
			mockClient: func() *mock.ConnectClientMock {
				mockConnectClient := &mock.ConnectClientMock{}
				mockConnectClient.On("GetVault", "test-id").Return(&onepassword.Vault{
					ID:          "test-id",
					Name:        "test-name",
					Description: "test-description",
				}, nil)
				return mockConnectClient
			},
			check: func(t *testing.T, vault *model.Vault, err error) {
				require.NoError(t, err)
				require.Equal(t, "test-id", vault.ID)
				require.Equal(t, "test-name", vault.Title)
				require.Equal(t, "test-description", vault.Description)
			},
		},
		"should return an error": {
			mockClient: func() *mock.ConnectClientMock {
				mockConnectClient := &mock.ConnectClientMock{}
				mockConnectClient.On("GetVault", "test-id").Return((*onepassword.Vault)(nil), errors.New("error"))
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
			client := &Client{connectClient: tc.mockClient()}
			vault, err := client.GetVault(context.Background(), "test-id")
			tc.check(t, vault, err)
		})
	}
}

func TestClient_GetVaultsByTitle(t *testing.T) {
	testCases := map[string]struct {
		mockClient func() *mock.ConnectClientMock
		check      func(t *testing.T, vaults []model.Vault, err error)
	}{
		"should return a single vault": {
			mockClient: func() *mock.ConnectClientMock {
				mockConnectClient := &mock.ConnectClientMock{}
				mockConnectClient.On("GetVaultsByTitle", "test-title").Return([]onepassword.Vault{
					{
						ID:          "test-id",
						Name:        "test-name",
						Description: "test-description",
					},
				}, nil)
				return mockConnectClient
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
			mockClient: func() *mock.ConnectClientMock {
				mockConnectClient := &mock.ConnectClientMock{}
				mockConnectClient.On("GetVaultsByTitle", "test-title").Return([]onepassword.Vault{
					{
						ID:          "test-id",
						Name:        "test-name",
						Description: "test-description",
					},
					{
						ID:          "test-id-2",
						Name:        "test-name-2",
						Description: "test-description-2",
					},
				}, nil)
				return mockConnectClient
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
			mockClient: func() *mock.ConnectClientMock {
				mockConnectClient := &mock.ConnectClientMock{}
				mockConnectClient.On("GetVaultsByTitle", "test-title").Return([]onepassword.Vault{}, errors.New("error"))
				return mockConnectClient
			},
			check: func(t *testing.T, vaults []model.Vault, err error) {
				require.Error(t, err)
				require.Empty(t, vaults)
			},
		},
	}

	for description, tc := range testCases {
		t.Run(description, func(t *testing.T) {
			client := &Client{connectClient: tc.mockClient()}
			vault, err := client.GetVaultsByTitle(context.Background(), "test-title")
			tc.check(t, vault, err)
		})
	}
}
