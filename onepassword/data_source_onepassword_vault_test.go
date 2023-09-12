package onepassword

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceOnePasswordVaultRead(t *testing.T) {
	expectedVault := onepassword.Vault{
		ID:          "vault-uuid",
		Name:        "Name of the vault",
		Description: "This vault will be retrieve",
	}

	var connectErr = errors.New("some request error")

	cases := map[string]struct {
		input              map[string]string
		getVaultRes        onepassword.Vault
		getVaultErr        error
		getVaultByTitleRes []onepassword.Vault
		getVaultByTitleErr error
		expected           onepassword.Vault
		expectedErr        error
	}{
		"by name": {
			input: map[string]string{
				"name": expectedVault.Name,
			},
			getVaultByTitleRes: []onepassword.Vault{
				expectedVault,
			},
			expected: expectedVault,
		},
		"by error": {
			input: map[string]string{
				"name": expectedVault.Name,
			},
			getVaultByTitleErr: connectErr,
			expectedErr:        connectErr,
		},
		"not_found_by_name": {
			input: map[string]string{
				"name": expectedVault.Name,
			},
			getVaultByTitleRes: []onepassword.Vault{},
			expectedErr:        fmt.Errorf("no vault found with name '%s'", expectedVault.Name),
		},
		"multiple_found_by_name": {
			input: map[string]string{
				"name": expectedVault.Name,
			},
			getVaultByTitleRes: []onepassword.Vault{
				expectedVault,
				expectedVault,
			},
			expectedErr: fmt.Errorf("multiple vaults found with name '%s'", expectedVault.Name),
		},
		"by uuid": {
			input: map[string]string{
				"uuid": expectedVault.ID,
			},
			getVaultRes: expectedVault,
			expected:    expectedVault,
		},
		"by uuid error": {
			input: map[string]string{
				"uuid": expectedVault.ID,
			},
			getVaultErr: connectErr,
			expectedErr: connectErr,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &testClient{
				GetVaultsByTitleFunc: func(title string) ([]onepassword.Vault, error) {
					return tc.getVaultByTitleRes, tc.getVaultByTitleErr
				},
				GetVaultFunc: func(uuid string) (*onepassword.Vault, error) {
					return &tc.getVaultRes, tc.getVaultErr
				},
			}
			dataSourceData := schema.TestResourceDataRaw(t, dataSourceOnepasswordVault().Schema, nil)

			for key, value := range tc.input {
				dataSourceData.Set(key, value)
			}

			err := dataSourceOnepasswordVaultRead(context.Background(), dataSourceData, client)

			if tc.expectedErr != nil {
				if err == nil || getErrorFromDiag(err) != tc.expectedErr.Error() {
					t.Errorf("Unexpected error occured. Expected %v, got %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Got unexpected error: %v", err)
				}
				assertResourceValue(t, dataSourceData, "uuid", tc.expected.ID)
				assertResourceValue(t, dataSourceData, "name", tc.expected.Name)
				assertResourceValue(t, dataSourceData, "description", tc.expected.Description)
			}
		})
	}
}

func assertResourceValue(t *testing.T, data *schema.ResourceData, key, expectedValue string) {
	value := data.Get(key)
	if value != expectedValue {
		t.Errorf("unexpected value for field %s. Expected %s, got %s", key, expectedValue, value)
	}
}

func getErrorFromDiag(d diag.Diagnostics) string {
	for _, dd := range d {
		if dd.Severity == diag.Error {
			return dd.Summary
		}
	}
	return ""
}
