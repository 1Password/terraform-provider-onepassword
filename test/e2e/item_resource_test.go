package integration

import (
	"fmt"
	"maps"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
)

var testItemsToCreate = map[op.ItemCategory]testItem{
	op.Login: {
		Attrs: map[string]string{
			"title":      "Test Login Create",
			"category":   "login",
			"username":   "testuser@example.com",
			"url":        "https://example.com",
			"note_value": "Test login note",
			"tags":       "testTag",
		},
	},
	op.Password: {
		Attrs: map[string]string{
			"title":    "Test Password Create",
			"category": "password",
			"tags":     "testTag",
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
			"tags":     "testTag",
		},
	},
	op.SecureNote: {
		Attrs: map[string]string{
			"title":      "Test Secure Note Create",
			"category":   "secure_note",
			"note_value": "This is a test secure note",
			"tags":       "testTag",
		},
	},
}

var testItemsUpdatedAttrs = map[op.ItemCategory]map[string]string{
	op.Login: {
		"title":      "Test Login Create - Updated",
		"username":   "updateduser@example.com",
		"password":   "updatedPassword",
		"url":        "https://updated-example.com",
		"note_value": "Updated login note",
		"tags":       "updatedTag",
	},
	op.Password: {
		"title":    "Test Password Create - Updated",
		"password": "updatedPassword",
		"tags":     "updatedTag",
	},
	op.Database: {
		"title":    "Test Database Create - Updated",
		"username": "updatedUsername",
		"password": "updatedPassword",
		"database": "updatedDatabase",
		"port":     "5432",
		"type":     "postgresql",
		"tags":     "updatedTag",
	},
	op.SecureNote: {
		"title":      "Test Secure Note Create - Updated",
		"note_value": "This is an updated secure note",
		"tags":       "updatedTag",
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

			// Determine if password_recipe is supported for this category
			// Only Login and Password support password_recipe currently
			usePasswordRecipe := tc.category == op.Login || tc.category == op.Password

			// Configs for creating and updating items
			initialConfig := maps.Clone(item.Attrs)
			updatedConfig := maps.Clone(item.Attrs)
			maps.Copy(updatedConfig, testItemsUpdatedAttrs[tc.category])

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, initialConfig, usePasswordRecipe),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "CREATE"),
						}, buildItemChecks("onepassword_item.test_item", initialConfig)...)...),
					},
					// Read/Import new item and verify it matches state
					{
						ResourceName:      "onepassword_item.test_item",
						ImportState:       true,
						ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, item.Attrs["title"]),
						ImportStateVerify: true,
						ImportStateVerifyIgnore: []string{
							"password_recipe",
						},
						ImportStateCheck: func(states []*terraform.InstanceState) error {
							t.Log("READ")
							return nil
						},
					},
					// Update new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, updatedConfig, false),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "UPDATE"),
						}, buildItemChecks("onepassword_item.test_item", updatedConfig)...)...),
					},
					// Delete new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemDataSourceConfig(
								map[string]string{
									"vault": testVaultID,
									"title": updatedConfig["title"],
								},
							),
						),
						ExpectError: regexp.MustCompile("Unable to read item"),
						Check:       logStep(t, "DELETE"),
					},
				},
			})
		})
	}
}

func logStep(t *testing.T, step string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Log(step)
		return nil
	}
}

func buildItemChecks(resourceName string, attrs map[string]string) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "uuid"),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
	}

	category := attrs["category"]
	if category == "login" || category == "password" || category == "database" {
		checks = append(checks, resource.TestCheckResourceAttrSet(resourceName, "password"))
	}

	for attr, expectedValue := range attrs {
		if attr == "tags" {
			checks = append(checks,
				resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "tags.0", expectedValue),
			)
			continue
		}

		checks = append(checks, resource.TestCheckResourceAttr(resourceName, attr, expectedValue))
	}

	return checks
}
