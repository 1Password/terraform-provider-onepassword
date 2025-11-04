package integration

import (
	"fmt"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/checks"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/password"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/uuid"
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
		"title":      "Test Login Create",
		"category":   "login",
		"username":   "updateduser@example.com",
		"password":   "updatedPassword",
		"url":        "https://updated-example.com",
		"note_value": "Updated login note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.Password: {
		"title":      "Test Password Create",
		"category":   "password",
		"password":   "updatedPassword",
		"note_value": "Updated password note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.Database: {
		"title":      "Test Database Create",
		"category":   "database",
		"username":   "updatedUsername",
		"password":   "updatedPassword",
		"database":   "updatedDatabase",
		"port":       "5432",
		"type":       "postgresql",
		"note_value": "Updated database note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	op.SecureNote: {
		"title":      "Test Secure Note Create",
		"category":   "secure_note",
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
			var itemUUID string

			// Build check functions for create step
			createChecks := []resource.TestCheckFunc{
				logStep(t, "CREATE"),
				uuid.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
			}
			bcCreate := checks.BuildItemChecks("onepassword_item.test_item", item.Attrs)
			createChecks = append(createChecks, bcCreate...)

			// Build checks for update step
			updateChecks := []resource.TestCheckFunc{
				logStep(t, "UPDATE"),
				uuid.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
			}
			bcUpdate := checks.BuildItemChecks("onepassword_item.test_item", testItemsUpdatedAttrs[tc.category])
			updateChecks = append(updateChecks, bcUpdate...)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, item.Attrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(createChecks...),
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
						Check: resource.ComposeAggregateTestCheckFunc(updateChecks...),
					},
					// Delete new item
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemDataSourceConfig(
								map[string]string{
									"vault": testVaultID,
									"title": fmt.Sprintf("%v", testItemsUpdatedAttrs[tc.category]["title"]),
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
	Bool := func(v bool) *bool { return &v }
	Int := func(v int) *int { return &v }

	testCases := []struct {
		name   string
		recipe password.PasswordRecipe
	}{
		{name: "Length32", recipe: password.PasswordRecipe{Length: Int(32), Letters: Bool(true), Digits: Bool(false), Symbols: Bool(false)}},
		{name: "Length16", recipe: password.PasswordRecipe{Length: Int(16), Letters: Bool(true), Digits: Bool(false), Symbols: Bool(false)}},
		{name: "WithSymbols", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(true), Digits: Bool(false), Letters: Bool(false)}},
		{name: "WithoutSymbols", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(false), Digits: Bool(true), Letters: Bool(true)}},
		{name: "WithDigits", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(false), Digits: Bool(true), Letters: Bool(false)}},
		{name: "WithoutDigits", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(true), Digits: Bool(false), Letters: Bool(true)}},
		{name: "WithLetters", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(false), Digits: Bool(false), Letters: Bool(true)}},
		{name: "WithoutLetters", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(true), Digits: Bool(true), Letters: Bool(false)}},
		{name: "AllCharacterTypesDisabled", recipe: password.PasswordRecipe{Length: Int(20), Symbols: Bool(false), Digits: Bool(false), Letters: Bool(false)}},
		{name: "LengthOnly", recipe: password.PasswordRecipe{Length: Int(20)}},
		{name: "InvalidLength0", recipe: password.PasswordRecipe{Length: Int(0)}},
		{name: "AllDefaults", recipe: password.PasswordRecipe{}},
	}

	// Test both Login and Password items
	items := []op.ItemCategory{op.Login, op.Password}

	for _, tc := range testCases {
		for _, item := range items {
			item := testItemsToCreate[item]

			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {
				recipeMap := password.BuildPasswordRecipeMap(tc.recipe)

				attrs := map[string]any{
					"title":           item.Attrs["title"],
					"category":        item.Attrs["category"],
					"password_recipe": recipeMap,
				}

				testStep := resource.TestStep{
					Config: tfconfig.CreateConfigBuilder()(
						tfconfig.ProviderConfig(),
						tfconfig.ItemResourceConfig(testVaultID, attrs),
					),
				}

				if tc.recipe.Length != nil && *tc.recipe.Length == 0 {
					testStep.ExpectError = regexp.MustCompile(`length value must be between 1 and 64`)
				} else {
					checks := password.BuildPasswordRecipeChecks("onepassword_item.test_item", recipeMap)
					testStep.Check = resource.ComposeAggregateTestCheckFunc(checks...)
				}

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps:                    []resource.TestStep{testStep},
				})
			})
		}
	}
}

func TestAccItemResourceSectionsAndFields(t *testing.T) {
	testCases := []struct {
		name        string
		createAttrs map[string]any
		updateAttrs map[string]any
	}{
		{
			name: "CreateSection",
			createAttrs: map[string]any{
				"section": []map[string]any{
					{
						"label": "Test Section",
					},
				},
			},
			updateAttrs: map[string]any{
				"section": []map[string]any{
					{
						"label": "Updated Section Label",
					},
					{
						"label": "Updated Section Label 2",
					},
				},
			},
		},
		{
			name: "CreateSectionWithField",
			createAttrs: map[string]any{
				"section": []map[string]any{
					{
						"label": "Test Section",
						"field": []map[string]any{
							{
								"label": "Test Field",
								"value": "2025-10-31",
								"type":  "DATE",
							},
						},
					},
				},
			},
			updateAttrs: map[string]any{
				"section": []map[string]any{
					{
						"label": "Test Section",
						"field": []map[string]any{
							{
								"label": "Updated Field",
								"value": "Test string",
								"type":  "STRING",
							},
							{
								"label": "Updated Field 2",
								"value": "2026-12-25",
								"type":  "DATE",
							},
						},
					},
				},
			},
		},
	}

	items := []op.ItemCategory{op.Login}

	for _, item := range items {
		item := testItemsToCreate[item]
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {
				var itemUUID string

				createAttrs := map[string]any{
					"title":    item.Attrs["title"],
					"category": item.Attrs["category"],
					"section":  tc.createAttrs["section"],
				}

				updateAttrs := map[string]any{
					"title":    item.Attrs["title"],
					"category": item.Attrs["category"],
					"section":  tc.updateAttrs["section"],
				}

				// Build check functions for create step
				createChecks := []resource.TestCheckFunc{
					logStep(t, "CREATE"),
					uuid.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
				}
				createChecks = append(createChecks, checks.BuildItemChecks("onepassword_item.test_item", createAttrs)...)

				// Build check functions for update step
				updateChecks := []resource.TestCheckFunc{
					logStep(t, "UPDATE"),
					uuid.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
				}
				updateChecks = append(updateChecks, checks.BuildItemChecks("onepassword_item.test_item", updateAttrs)...)

				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						// Create new item
						{
							Config: tfconfig.CreateConfigBuilder()(
								tfconfig.ProviderConfig(),
								tfconfig.ItemResourceConfig(testVaultID, createAttrs),
							),
							Check: resource.ComposeAggregateTestCheckFunc(createChecks...),
						},
						// Update new item
						{
							Config: tfconfig.CreateConfigBuilder()(
								tfconfig.ProviderConfig(),
								tfconfig.ItemResourceConfig(testVaultID, updateAttrs),
							),
							Check: resource.ComposeAggregateTestCheckFunc(updateChecks...),
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
