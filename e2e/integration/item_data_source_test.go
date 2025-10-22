package integration

import (
	"regexp"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/e2e/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testVaultID = "t7dnwbjh6nlyw475wl3m442sdi"

type testItem struct {
	Title string
	UUID  string
	Attrs map[string]string
}

var testItems = map[string]testItem{
	"Login": {
		Title: "Test Login",
		UUID:  "dsrwv5dyacw4f7pdrfnmh36pne",
		Attrs: map[string]string{
			"category": "login",
			"username": "testUsername",
			"password": "testPassword",
			"url":      "www.example.com",
		},
	},
	"Password": {
		Title: "Test Password",
		UUID:  "nlinya3ju5lagllswd6ggleoqi",
		Attrs: map[string]string{
			"category": "password",
			"password": "samplePassword",
		},
	},
	"Database": {
		Title: "Test Database",
		UUID:  "cq24ebmitcdwpt52f4xqdrq3ce",
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
		UUID:  "culehzcmv2qcc62qjsngj5ghyi",
		Attrs: map[string]string{
			"category":   "secure_note",
			"note_value": "Test note",
		},
	},
}

func TestAccItemDataSource(t *testing.T) {
	config, err := utils.GetTestConfig()
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
				checks = append(checks, utils.ValidateResourceAttribute("data.onepassword_item.test", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: utils.TestAccItemDataSourceConfig(config, testVaultID, tc.identifierType, identifierValue),
					Check:  resource.ComposeAggregateTestCheckFunc(checks...),
				}},
			})
		})
	}
}

func TestAccItemDataSource_NotFound(t *testing.T) {
	config, err := utils.GetTestConfig()
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
					Config:      utils.TestAccItemDataSourceConfig(config, testVaultID, tc.lookupType, tc.lookupValue),
					ExpectError: regexp.MustCompile(`Unable to read item`),
				}},
			})
		})
	}
}
