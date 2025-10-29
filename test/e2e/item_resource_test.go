package integration

import (
	"fmt"
	"maps"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testItemsToCreate = map[op.ItemCategory]testItem{
	op.Login: {
		Attrs: map[string]string{
			"title":      "Test Login Create",
			"category":   "login",
			"username":   "testuser@example.com",
			"password":   "testPassword",
			"url":        "https://example.com",
			"note_value": "Test login note",
		},
	},
	op.Password: {
		Attrs: map[string]string{
			"title":    "Test Password Create",
			"category": "password",
			"password": "testPassword",
		},
	},
	op.Database: {
		Attrs: map[string]string{
			"title":    "Test Database Create",
			"category": "database",
			"username": "testUsername",
			"password": "testPassword",
			"database": "testDatabase",
			"port":     "3306",
			"type":     "mysql",
		},
	},
	op.SecureNote: {
		Attrs: map[string]string{
			"title":      "Test Secure Note Create",
			"category":   "secure_note",
			"note_value": "This is a test secure note",
		},
	},
}

var testItemsUpdatedAttrs = map[op.ItemCategory]map[string]string{
	op.Login: {
		"username":   "updateduser@example.com",
		"password":   "updatedPassword",
		"url":        "https://updated-example.com",
		"note_value": "Updated login note",
	},
	op.Password: {
		"password": "updatedPassword",
	},
	op.Database: {
		"username": "updatedUsername",
		"password": "updatedPassword",
		"database": "updatedDatabase",
		"port":     "5432",
		"type":     "postgresql",
	},
	op.SecureNote: {
		"note_value": "This is an updated secure note",
	},
}

func TestAccItemResourceCRUD(t *testing.T) {
	testCases := []struct {
		category op.ItemCategory
		name     string
	}{
		{op.Login, "Login"},
		{op.Password, "Password"},
		{op.Database, "Database"},
		{op.SecureNote, "SecureNote"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := testItemsToCreate[tc.category]

			// Create Config
			initialConfig := maps.Clone(item.Attrs)

			// Update Config
			updatedConfig := maps.Clone(item.Attrs)
			maps.Copy(updatedConfig, testItemsUpdatedAttrs[tc.category])

			// Create Checks
			initialChecks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Attrs["title"]),
			}
			for attr, expectedValue := range item.Attrs {
				initialChecks = append(initialChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
			}

			// Update Checks
			updatedChecks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Attrs["title"]),
			}
			for attr, expectedValue := range testItemsUpdatedAttrs[tc.category] {
				updatedChecks = append(updatedChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, initialConfig),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							resource.TestCheckFunc(func(s *terraform.State) error {
								t.Logf("CREATE")
								return nil
							}),
						}, initialChecks...)...),
					},
					// Read
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, initialConfig),
						),
						ResourceName:  "onepassword_item.test_item",
						ImportStateId: fmt.Sprintf("vaults/%s/items/%s", "t7dnwbjh6nlyw475wl3m442sdi", item.Title),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							resource.TestCheckFunc(func(s *terraform.State) error {
								t.Logf("READ")
								return nil
							}),
						}, initialChecks...)...),
					},
					// Update
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, updatedConfig),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							resource.TestCheckFunc(func(s *terraform.State) error {
								t.Logf("UPDATE")
								return nil
							}),
						}, updatedChecks...)...),
					},
					// Delete
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
						),
						Check: resource.TestCheckFunc(func(s *terraform.State) error {
							t.Logf("DELETE")
							return nil
						}),
					},
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemDataSourceConfig(
								map[string]string{
									"vault": "t7dnwbjh6nlyw475wl3m442sdi",
									"title": item.Title,
								},
							),
						),
						ExpectError: regexp.MustCompile("Unable to read item"),
					},
				},
			})
		})
	}
}
