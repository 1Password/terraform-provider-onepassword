package integration

import (
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
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
			"category": "document",
			"file.0.name":     "test.txt",
			"file.0.content":  "This is a test\n",
			"file.0.content_base64": "VGhpcyBpcyBhIHRlc3QK",
		},
	},
}

func TestAccItemDataSource(t *testing.T) {
	config, err := config.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	testCases := []struct {
		name           string
		item           testItem
		identifierType string
	}{
		{"LoginByTitle", testItems[op.Login], "title"},
		{"LoginByUUID", testItems[op.Login], "uuid"},
		{"PasswordByTitle", testItems[op.Password], "title"},
		{"PasswordByUUID", testItems[op.Password], "uuid"},
		{"DatabaseByTitle", testItems[op.Database], "title"},
		{"DatabaseByUUID", testItems[op.Database], "uuid"},
		{"SecureNoteByTitle", testItems[op.SecureNote], "title"},
		{"SecureNoteByUUID", testItems[op.SecureNote], "uuid"},
		{"DocumentByTitle", testItems[op.Document], "title"},
		{"DocumentByUUID", testItems[op.Document], "uuid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			identifierValue := tc.item.Title
			if tc.identifierType == "uuid" {
				identifierValue = tc.item.UUID
			}

			checks := make([]resource.TestCheckFunc, 0, len(tc.item.Attrs))
			for attr, expectedValue := range tc.item.Attrs {
				checks = append(checks, resource.TestCheckResourceAttr("data.onepassword_item.test", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: tfconfig.DataSource(tfconfig.DataSourceConfigParams{
						TestConfig:      config,
						DataSource:  "onepassword_item",
						Vault:           testVaultID,
						IdentifierType:  tc.identifierType,
						IdentifierValue: identifierValue,
					}),
					Check:  resource.ComposeAggregateTestCheckFunc(checks...),
				}},
			})
		})
	}
}

func TestAccItemDataSource_NotFound(t *testing.T) {
	config, err := config.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	testCases := []struct {
		name        string
		identifierType  string
		identifierValue string
	}{
		{"ByTitle", "title", "invalid-title"},
		{"ByUUID", "uuid", "invalid-uuid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: tfconfig.DataSource(tfconfig.DataSourceConfigParams{
						TestConfig:      config,
						DataSource:  "onepassword_item",
						Vault:           testVaultID,
						IdentifierType:  tc.identifierType,
						IdentifierValue: tc.identifierValue,
					}),
					ExpectError: regexp.MustCompile(`Unable to read item`),
				}},
			})
		})
	}
}
