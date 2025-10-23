package integration

import (
	"regexp"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testVaultID = "bbucuyq2nn4fozygwttxwizpcy"

type testItem struct {
	Title string
	UUID  string
	Attrs map[string]string
}

var testItems = map[string]testItem{
	"Login": {
		Title: "Test Login",
		UUID:  "5axoqbjhbx3u7wqmersrg6qnqy",
		Attrs: map[string]string{
			"category": "login",
			"username": "testUsername",
			"password": "testPassword",
			"url":      "www.example.com",
		},
	},
	"Password": {
		Title: "Test Password",
		UUID:  "axoqeauq7ilndgdpimb4j4dwhi",
		Attrs: map[string]string{
			"category": "password",
			"password": "testPassword",
		},
	},
	"Database": {
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
	"SecureNote": {
		Title: "Test Secure Note",
		UUID:  "5xbca3eblv5kxkszrbuhdame4a",
		Attrs: map[string]string{
			"category":   "secure_note",
			"note_value": "This is a test secure note for terraform-provider-onepassword",
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
		{"LoginByTitle", testItems["Login"], "title"},
		{"LoginByUUID", testItems["Login"], "uuid"},
		{"PasswordByTitle", testItems["Password"], "title"},
		{"PasswordByUUID", testItems["Password"], "uuid"},
		{"DatabaseByTitle", testItems["Database"], "title"},
		{"DatabaseByUUID", testItems["Database"], "uuid"},
		{"SecureNoteByTitle", testItems["SecureNote"], "title"},
		{"SecureNoteByUUID", testItems["SecureNote"], "uuid"},
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
					Config: tfconfig.ItemDataSourceConfig(config, testVaultID, tc.identifierType, identifierValue),
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
		lookupType  string
		lookupValue string
	}{
		{"ByTitle", "title", "invalid-title"},
		{"ByUUID", "uuid", "invalid-uuid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      tfconfig.ItemDataSourceConfig(config, testVaultID, tc.lookupType, tc.lookupValue),
					ExpectError: regexp.MustCompile(`Unable to read item`),
				}},
			})
		})
	}
}
