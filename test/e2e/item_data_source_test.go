package integration

import (
	"fmt"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils"
)

const testVaultID = "bbucuyq2nn4fozygwttxwizpcy"

type testItem struct {
	Title string
	UUID  string
	Attrs map[string]string
}

var testItems = map[op.ItemCategory]testItem{
	op.Login: {
		Title: "Test Login",
		UUID:  "5axoqbjhbx3u7wqmersrg6qnqy",
		Attrs: map[string]string{
			"category": "login",
			"username": "testUsername",
			"password": "testPassword",
			"url":      "www.example.com",
		},
	},
	op.Password: {
		Title: "Test Password",
		UUID:  "axoqeauq7ilndgdpimb4j4dwhi",
		Attrs: map[string]string{
			"category": "password",
			"password": "testPassword",
		},
	},
	op.Database: {
		Title: "Test Database",
		UUID:  "ck6mbmf3yjps6gk5qldnx4frni",
		Attrs: map[string]string{
			"category": "database",
			"username": "testUsername",
			"password": "testPassword",
			"database": "testDatabase",
			"port":     "3306",
			"type":     "mysql",
		},
	},
	op.SecureNote: {
		Title: "Test Secure Note",
		UUID:  "5xbca3eblv5kxkszrbuhdame4a",
		Attrs: map[string]string{
			"category":   "secure_note",
			"note_value": "This is a test secure note for terraform-provider-onepassword",
		},
	},
	op.Document: {
		Title: "Test Document",
		UUID:  "p6uyugpmxo6zcxo5fdfctet7xa",
		Attrs: map[string]string{
			"category":              "document",
			"file.0.name":           "test.txt",
			"file.0.content":        "This is a test\n",
			"file.0.content_base64": "VGhpcyBpcyBhIHRlc3QK",
		},
	},
	op.SSHKey: {
		Title: "Test SSH Key",
		UUID:  "5dbnxvhcknslz4mcaz7lobzt6i",
		Attrs: map[string]string{
			"category": "ssh_key",
		},
	},
}

func TestAccItemDataSource(t *testing.T) {
	serviceAccountToken, err := config.GetServiceAccountToken()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	testCases := []struct {
		name                 string
		item                 testItem
		itemDataSourceConfig tfconfig.ItemDataSource
	}{
		{
			name: "LoginByTitle",
			item: testItems[op.Login],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.Login].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "LoginByUUID",
			item: testItems[op.Login],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.Login].UUID,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "PasswordByTitle",
			item: testItems[op.Password],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.Password].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "PasswordByUUID",
			item: testItems[op.Password],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.Password].UUID,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "DatabaseByTitle",
			item: testItems[op.Database],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.Database].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "DatabaseByUUID",
			item: testItems[op.Database],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.Database].UUID,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "SecureNoteByTitle",
			item: testItems[op.SecureNote],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.SecureNote].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "SecureNoteByUUID",
			item: testItems[op.SecureNote],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.SecureNote].UUID,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "DocumentByTitle",
			item: testItems[op.Document],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.Document].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "DocumentByUUID",
			item: testItems[op.Document],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.Document].UUID,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "SSHKeyByTitle",
			item: testItems[op.SSHKey],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": testItems[op.SSHKey].Title,
					"vault": testVaultID,
				},
			},
		},
		{
			name: "SSHKeyByUUID",
			item: testItems[op.SSHKey],
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  testItems[op.SSHKey].UUID,
					"vault": testVaultID,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataSourceBuilder := tfconfig.CreateItemDataSourceConfigBuilder()

			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "title", tc.item.Title),
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "uuid", tc.item.UUID),
			}

			for attr, expectedValue := range tc.item.Attrs {
				checks = append(checks, resource.TestCheckResourceAttr("data.onepassword_item.test_item", attr, expectedValue))
			}

			// Validate SSH keys
			if tc.item.Attrs["category"] == "ssh_key" {
				checks = append(checks, resource.TestCheckFunc(func(s *terraform.State) error {
					item, ok := s.RootModule().Resources["data.onepassword_item.test_item"]
					if !ok {
						return fmt.Errorf("resource not found in state")
					}

					publicKey := item.Primary.Attributes["public_key"]
					privateKey := item.Primary.Attributes["private_key"]

					return utils.ValidateSSHKeys(publicKey, privateKey)
				}))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderAuthWithServiceAccount(tc.itemDataSourceConfig.Auth),
						tfconfig.ItemDataSourceConfig(tc.itemDataSourceConfig.Params),
					),
					Check: resource.ComposeAggregateTestCheckFunc(checks...),
				}},
			})
		})
	}
}

func TestAccItemDataSource_NotFound(t *testing.T) {
	serviceAccountToken, err := config.GetServiceAccountToken()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	testCases := []struct {
		name                 string
		item                 testItem
		itemDataSourceConfig tfconfig.ItemDataSource
	}{
		{
			name: "ByTitle",
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"title": "invalid-title",
					"vault": testVaultID,
				},
			},
		},
		{
			name: "ByUUID",
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid":  "invalid-uuid",
					"vault": testVaultID,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataSourceBuilder := tfconfig.CreateItemDataSourceConfigBuilder()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderAuthWithServiceAccount(tc.itemDataSourceConfig.Auth),
						tfconfig.ItemDataSourceConfig(tc.itemDataSourceConfig.Params),
					),
					ExpectError: regexp.MustCompile(`Unable to read item`),
				}},
			})
		})
	}
}
