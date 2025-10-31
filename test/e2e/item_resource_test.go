package integration

import (
	"fmt"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/validate"
)

type testResourceItem struct {
	Attrs map[string]any
}

var testItemsToCreate = map[op.ItemCategory]testResourceItem{
	op.Login: {
		Attrs: map[string]any{
			"title":      "Test Login Create",
			"category":   "login",
			"username":   "testuser@example.com",
			"password":   "testPassword",
			"url":        "https://example.com",
			"note_value": "Test login note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
	op.Password: {
		Attrs: map[string]any{
			"title":      "Test Password Create",
			"category":   "password",
			"password":   "testPassword",
			"note_value": "Test password note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
	op.Database: {
		Attrs: map[string]any{
			"title":      "Test Database Create",
			"category":   "database",
			"username":   "testUsername",
			"password":   "testPassword",
			"database":   "testDatabase",
			"port":       "3306",
			"type":       "mysql",
			"note_value": "Test database note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
	op.SecureNote: {
		Attrs: map[string]any{
			"title":      "Test Secure Note Create",
			"category":   "secure_note",
			"note_value": "This is a test secure note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
}

var testItemsUpdatedAttrs = map[op.ItemCategory]map[string]any{
	op.Login: {
		"title":      "Test Login Create - Updated",
		"username":   "updateduser@example.com",
		"password":   "updatedPassword",
		"url":        "https://updated-example.com",
		"note_value": "Updated login note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.Password: {
		"title":      "Test Password Create - Updated",
		"password":   "updatedPassword",
		"note_value": "Updated password note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.Database: {
		"title":      "Test Database Create - Updated",
		"username":   "updatedUsername",
		"password":   "updatedPassword",
		"database":   "updatedDatabase",
		"port":       "5432",
		"type":       "postgresql",
		"note_value": "Updated database note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.SecureNote: {
		"title":      "Test Secure Note Create - Updated",
		"note_value": "This is an updated secure note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
}

func TestAccItemResource(t *testing.T) {
	testCases := []struct {
		category op.ItemCategory
		name     string
	}{
		{category: op.Login, name: "Login"},
		{category: op.Password, name: "Password"},
		{category: op.Database, name: "Database"},
		{category: op.SecureNote, name: "SecureNote"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := testItemsToCreate[tc.category]

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, item.Attrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "CREATE"),
						}, validate.BuildItemChecks("onepassword_item.test_item", item.Attrs)...)...),
					},
					// Read/Import new item and verify it matches state
					{
						ResourceName:      "onepassword_item.test_item",
						ImportState:       true,
						ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, item.Attrs["title"]),
						ImportStateVerify: true,
						ImportStateCheck: func(states []*terraform.InstanceState) error {
							t.Log("READ")
							return nil
						},
					},
					// Update new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, testItemsUpdatedAttrs[tc.category]),
						),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							logStep(t, "UPDATE"),
						}, validate.BuildItemChecks("onepassword_item.test_item", testItemsUpdatedAttrs[tc.category])...)...),
					},
					// Delete new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemDataSourceConfig(
								map[string]string{
									"vault": testVaultID,
									"title": fmt.Sprintf("%v", testItemsUpdatedAttrs[tc.category]["title"])},
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
		name   string
		recipe map[string]any
	}{
		{name: "Length32", recipe: map[string]any{"length": 32, "symbols": false, "digits": false, "letters": true}},
		{name: "Length16", recipe: map[string]any{"length": 16, "symbols": false, "digits": false, "letters": true}},
		{name: "WithSymbols", recipe: map[string]any{"length": 20, "symbols": true, "digits": false, "letters": false}},
		{name: "WithoutSymbols", recipe: map[string]any{"length": 20, "symbols": false, "digits": true, "letters": true}},
		{name: "WithDigits", recipe: map[string]any{"length": 20, "symbols": false, "digits": true, "letters": false}},
		{name: "WithoutDigits", recipe: map[string]any{"length": 20, "symbols": true, "digits": false, "letters": true}},
		{name: "WithLetters", recipe: map[string]any{"length": 20, "symbols": false, "digits": false, "letters": true}},
		{name: "WithoutLetters", recipe: map[string]any{"length": 20, "symbols": true, "digits": true, "letters": false}},
	}

	// Test both Login and Password items
	items := []op.ItemCategory{op.Login, op.Password}

	for _, item := range items {
		item := testItemsToCreate[item]

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {

				attrs := map[string]any{
					"title":           item.Attrs["title"],
					"category":        item.Attrs["category"],
					"password_recipe": tc.recipe,
				}

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: tfconfig.CreateConfigBuilder()(
								tfconfig.ProviderConfig(),
								tfconfig.ItemResourceConfig(testVaultID, attrs),
							),
							Check: resource.ComposeAggregateTestCheckFunc(validate.BuildPasswordRecipeChecks("onepassword_item.test_item", tc.recipe)...),
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
