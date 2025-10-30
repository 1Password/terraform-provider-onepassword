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
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/validate"
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
			"tags":       "testTag",
		},
	},
	op.Password: {
		Attrs: map[string]string{
			"title":    "Test Password Create",
			"category": "password",
			"password": "testPassword",
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

func TestAccItemResource(t *testing.T) {
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
							tfconfig.ItemResourceConfig(testVaultID, initialConfig, nil),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "CREATE"),
						}, validate.BuildItemChecks("onepassword_item.test_item", initialConfig)...)...),
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
							tfconfig.ItemResourceConfig(testVaultID, updatedConfig, nil),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "UPDATE"),
						}, validate.BuildItemChecks("onepassword_item.test_item", updatedConfig)...)...),
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

func TestAccItemResourcePasswordGeneration(t *testing.T) {
	testCases := []struct {
		name           string
		passwordRecipe *tfconfig.PasswordRecipe
	}{
		{"Length32", &tfconfig.PasswordRecipe{Length: 32, Symbols: false, Digits: false, Letters: true}},
		{"Length16", &tfconfig.PasswordRecipe{Length: 16, Symbols: false, Digits: false, Letters: true}},
		{"WithSymbols", &tfconfig.PasswordRecipe{Length: 20, Symbols: true, Digits: false, Letters: false}},
		{"WithoutSymbols", &tfconfig.PasswordRecipe{Length: 20, Symbols: false, Digits: true, Letters: true}},
		{"WithDigits", &tfconfig.PasswordRecipe{Length: 20, Symbols: false, Digits: true, Letters: false}},
		{"WithoutDigits", &tfconfig.PasswordRecipe{Length: 20, Symbols: true, Digits: false, Letters: true}},
		{"WithLetters", &tfconfig.PasswordRecipe{Length: 20, Symbols: false, Digits: false, Letters: true}},
		{"WithoutLetters", &tfconfig.PasswordRecipe{Length: 20, Symbols: true, Digits: true, Letters: false}},
	}

	// Test both Login and Password items
	items := []op.ItemCategory{op.Login, op.Password}

	for _, item := range items {
		item := testItemsToCreate[item]

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {

				initialConfig := maps.Clone(item.Attrs)

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: tfconfig.CreateConfigBuilder()(
								tfconfig.ProviderConfig(),
								tfconfig.ItemResourceConfig(testVaultID, initialConfig, tc.passwordRecipe),
							),
							Check: resource.ComposeAggregateTestCheckFunc(validate.BuildPasswordRecipeChecks("onepassword_item.test_item", tc.passwordRecipe)...),
						},
					},
				})
			})
		}
	}
}

// logStep logs the current test step for easier test debugging
func logStep(t *testing.T, step string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Log(step)
		return nil
	}
}
