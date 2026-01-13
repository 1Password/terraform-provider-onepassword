package integration

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/attributes"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/checks"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/cleanup"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/client"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/password"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/sections"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/uuid"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/vault"
)

type testResourceItem struct {
	Attrs map[string]any
}

var testItemsToCreate = map[model.ItemCategory]testResourceItem{
	model.Login: {
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
	model.Password: {
		Attrs: map[string]any{
			"title":      "Test Password Create",
			"category":   "password",
			"password":   "testPassword",
			"note_value": "Test password note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
	model.Database: {
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
	model.SecureNote: {
		Attrs: map[string]any{
			"title":      "Test Secure Note Create",
			"category":   "secure_note",
			"note_value": "This is a test secure note",
			"tags":       []string{"firstTestTag", "secondTestTag"},
		},
	},
}

var testItemsUpdatedAttrs = map[model.ItemCategory]map[string]any{
	model.Login: {
		"title":      "Test Login Create",
		"category":   "login",
		"username":   "updateduser@example.com",
		"password":   "updatedPassword",
		"url":        "https://updated-example.com",
		"note_value": "Updated login note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	model.Password: {
		"title":      "Test Password Create",
		"category":   "password",
		"password":   "updatedPassword",
		"note_value": "Updated password note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
	model.Database: {
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
	model.SecureNote: {
		"title":      "Test Secure Note Create",
		"category":   "secure_note",
		"note_value": "This is an updated secure note",
		"tags":       []string{"firstUpdatedTestTag", "secondUpdatedTestTag"},
	},
}

func TestAccItemResource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		category model.ItemCategory
		name     string
	}{
		{category: model.Login, name: "Login"},
		{category: model.Password, name: "Password"},
		{category: model.Database, name: "Database"},
		{category: model.SecureNote, name: "SecureNote"},
	}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate unique identifier for this test run to avoid conflicts in parallel execution
			uniqueID := uuid.New().String()

			item := testItemsToCreate[tc.category]
			// Create a copy of item attributes and update title with unique ID
			createAttrs := maps.Clone(item.Attrs)
			createAttrs["title"] = addUniqueIDToTitle(createAttrs["title"].(string), uniqueID)

			// Create a copy of updated attributes and update title with unique ID
			updatedAttrs := maps.Clone(testItemsUpdatedAttrs[tc.category])
			updatedAttrs["title"] = addUniqueIDToTitle(updatedAttrs["title"].(string), uniqueID)

			var itemUUID string

			// Build check functions for create step
			createChecks := []resource.TestCheckFunc{
				logStep(t, "CREATE"),
				uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
				cleanup.RegisterItem(t, &itemUUID, testVaultID),
			}
			bcCreate := checks.BuildItemChecks("onepassword_item.test_item", createAttrs)
			createChecks = append(createChecks, bcCreate...)

			// Build checks for update step
			updateChecks := []resource.TestCheckFunc{
				logStep(t, "UPDATE"),
				uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
			}
			bcUpdate := checks.BuildItemChecks("onepassword_item.test_item", updatedAttrs)
			updateChecks = append(updateChecks, bcUpdate...)

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
					// Read/Import new item and verify it matches state
					{
						ResourceName:      "onepassword_item.test_item",
						ImportState:       true,
						ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, createAttrs["title"]),
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
							tfconfig.ItemResourceConfig(testVaultID, updatedAttrs),
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
									"title": fmt.Sprintf("%v", updatedAttrs["title"]),
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
	t.Parallel()

	testCases := []struct {
		name   string
		recipe password.PasswordRecipe
	}{
		{name: "Length32", recipe: password.PasswordRecipe{Length: 32, Digits: false, Symbols: false}},
		{name: "Length16", recipe: password.PasswordRecipe{Length: 16, Digits: false, Symbols: false}},
		{name: "WithSymbols", recipe: password.PasswordRecipe{Length: 20, Digits: false, Symbols: true}},
		{name: "WithoutSymbols", recipe: password.PasswordRecipe{Length: 20, Symbols: false, Digits: true}},
		{name: "WithDigits", recipe: password.PasswordRecipe{Length: 20, Symbols: false, Digits: true}},
		{name: "WithoutDigits", recipe: password.PasswordRecipe{Length: 20, Symbols: true, Digits: false}},
		{name: "AllCharacterTypesDisabled", recipe: password.PasswordRecipe{Length: 20, Symbols: false, Digits: false}},
		{name: "InvalidLength0", recipe: password.PasswordRecipe{Length: 0}},
		{name: "InvalidLength65", recipe: password.PasswordRecipe{Length: 65}},
	}

	testVaultID := vault.GetTestVaultID(t)

	// Test both Login and Password items
	items := []model.ItemCategory{model.Login, model.Password}

	for _, tc := range testCases {
		for _, item := range items {
			item := testItemsToCreate[item]

			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {
				t.Parallel()

				// Generate unique identifier for this test run to avoid conflicts in parallel execution
				uniqueID := uuid.New().String()

				recipeMap := password.BuildPasswordRecipeMap(tc.recipe)

				attrs := map[string]any{
					"title":           addUniqueIDToTitle(item.Attrs["title"].(string), uniqueID),
					"category":        item.Attrs["category"],
					"password_recipe": recipeMap,
				}

				testStep := resource.TestStep{
					Config: tfconfig.CreateConfigBuilder()(
						tfconfig.ProviderConfig(),
						tfconfig.ItemResourceConfig(testVaultID, attrs),
					),
				}

				if tc.recipe.Length < 1 || tc.recipe.Length > 64 {
					testStep.ExpectError = regexp.MustCompile(`length value must be between 1 and 64`)
				} else {
					var itemUUID string
					checks := []resource.TestCheckFunc{
						uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
						cleanup.RegisterItem(t, &itemUUID, testVaultID),
					}
					checks = append(checks, password.BuildPasswordRecipeChecks("onepassword_item.test_item", tc.recipe)...)
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

// Test that letters is not supported and will error if configured as this field is deprecated
func TestAccItemResourcePasswordGeneration_InvalidLetters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		letters bool
	}{
		{name: "LettersTrue", letters: true},
		{name: "LettersFalse", letters: false},
	}

	testVaultID := vault.GetTestVaultID(t)

	item := testItemsToCreate[model.Login]

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate unique identifier for this test run to avoid conflicts in parallel execution
			uniqueID := uuid.New().String()

			recipeMap := map[string]any{
				"length":  20,
				"symbols": false,
				"digits":  false,
				"letters": tc.letters,
			}

			attrs := map[string]any{
				"title":           addUniqueIDToTitle(item.Attrs["title"].(string), uniqueID),
				"category":        item.Attrs["category"],
				"password_recipe": recipeMap,
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, attrs),
						),
						ExpectError: regexp.MustCompile(`An argument named "letters" is not expected here`),
					},
				},
			})
		})
	}
}

// TestAccItemResourceSectionFieldPasswordGeneration tests the generation of passwords on fields
func TestAccItemResourceSectionFieldPasswordGeneration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		recipe password.PasswordRecipe
	}{
		{name: "Length32", recipe: password.PasswordRecipe{Length: 32, Digits: false, Symbols: false}},
		{name: "WithSymbols", recipe: password.PasswordRecipe{Length: 20, Digits: false, Symbols: true}},
		{name: "WithDigits", recipe: password.PasswordRecipe{Length: 20, Symbols: false, Digits: true}},
		{name: "AllCharacterTypesDisabled", recipe: password.PasswordRecipe{Length: 20, Symbols: false, Digits: false}},
		{name: "InvalidLength", recipe: password.PasswordRecipe{Length: 0}},
	}

	testVaultID := vault.GetTestVaultID(t)

	item := testItemsToCreate[model.Login]

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Generate unique identifier for this test run to avoid conflicts in parallel execution
			uniqueID := uuid.New().String()

			recipeMap := password.BuildPasswordRecipeMap(tc.recipe)

			// Create a field with password recipe in a section
			testSection := sections.TestSection{
				Label: "Credentials",
				Fields: []sections.TestField{
					{
						Label:          "API Key",
						Type:           "CONCEALED",
						PasswordRecipe: &recipeMap,
					},
				},
			}

			attrs := map[string]any{
				"title":    addUniqueIDToTitle(item.Attrs["title"].(string), uniqueID),
				"category": item.Attrs["category"],
				"section":  sections.MapSections([]sections.TestSection{testSection}),
			}

			testStep := resource.TestStep{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
			}

			if tc.recipe.Length < 1 || tc.recipe.Length > 64 {
				testStep.ExpectError = regexp.MustCompile(`Invalid Attribute Value`)
			} else {
				var itemUUID string
				checks := []resource.TestCheckFunc{
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
				}
				checks = append(checks, password.BuildPasswordRecipeChecksForField("onepassword_item.test_item", "section.0.field.0", tc.recipe)...)
				testStep.Check = resource.ComposeAggregateTestCheckFunc(checks...)
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    []resource.TestStep{testStep},
			})
		})
	}
}

// TestAccItemResourceSectionList_ValueAndPasswordRecipeConflict tests that value and password_recipe
// cannot be specified together in section list fields:
func TestAccItemResourceSectionList_ValueAndPasswordRecipeConflict(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()

	recipeMap := map[string]any{
		"length":  20,
		"digits":  true,
		"symbols": true,
	}

	// Create section with a field that has both value and password_recipe set - this should fail validation
	testSection := sections.TestSection{
		Label: "Credentials",
		Fields: []sections.TestField{
			{
				Label:          "Conflicting Field",
				Type:           "CONCEALED",
				Value:          "my-explicit-value",
				PasswordRecipe: &recipeMap,
			},
		},
	}

	attrs := map[string]any{
		"title":    addUniqueIDToTitle("Test Section List Value Recipe Conflict", uniqueID),
		"category": "login",
		"section":  sections.MapSections([]sections.TestSection{testSection}),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func TestAccItemResourceSectionsAndFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		create sections.TestSectionData
		update sections.TestSectionData
	}{
		{
			name: "RemoveSection",
			create: sections.TestSectionData{
				Sections: []sections.TestSection{
					{Label: "Test Section 1"},
					{Label: "Test Section 2"},
				},
			},
			update: sections.TestSectionData{
				Sections: []sections.TestSection{
					{Label: "Test Section 1"},
				},
			},
		},
		{
			name: "RemoveFieldFromSection",
			create: sections.TestSectionData{
				Sections: []sections.TestSection{
					{
						Label: "Test Section",
						Fields: []sections.TestField{
							{Label: "Field 1", Value: "value1", Type: "STRING"},
							{Label: "Field 2", Value: "value2", Type: "STRING"},
						},
					},
				},
			},
			update: sections.TestSectionData{
				Sections: []sections.TestSection{
					{
						Label: "Test Section",
						Fields: []sections.TestField{
							{Label: "Field 1", Value: "value1", Type: "STRING"},
						},
					},
				},
			},
		},
		{
			name: "AddFieldToExistingSection",
			create: sections.TestSectionData{
				Sections: []sections.TestSection{
					{Label: "Test Section"},
				},
			},
			update: sections.TestSectionData{
				Sections: []sections.TestSection{
					{
						Label: "Test Section",
						Fields: []sections.TestField{
							{Label: "New Field", Value: "new value", Type: "STRING"},
						},
					},
				},
			},
		},
		{
			name: "MultipleSectionsWithMultipleFields",
			create: sections.TestSectionData{
				Sections: []sections.TestSection{
					{
						Label: "Personal Info",
						Fields: []sections.TestField{
							{Label: "Email", Value: "test@example.com", Type: "EMAIL"},
							{Label: "Date", Value: "1990-01-01", Type: "DATE"},
						},
					},
					{
						Label: "Additional Info",
						Fields: []sections.TestField{
							{Label: "Website", Value: "https://example.com", Type: "URL"},
							{Label: "Concealed Field", Value: "secret", Type: "CONCEALED"},
						},
					},
				},
			},
			update: sections.TestSectionData{
				Sections: []sections.TestSection{
					{
						Label: "Personal Info",
						Fields: []sections.TestField{
							{Label: "Updated Email", Value: "updated@example.com", Type: "EMAIL"},
							{Label: "Date", Value: "1990-01-01", Type: "DATE"},
						},
					},
					{
						Label: "Additional Info",
						Fields: []sections.TestField{
							{Label: "Website", Value: "https://updated.com", Type: "URL"},
							{Label: "Concealed Field", Value: "secret", Type: "CONCEALED"},
							{Label: "Notes", Value: "Some notes", Type: "STRING"},
						},
					},
				},
			},
		},
	}

	items := []model.ItemCategory{model.Login}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		for _, item := range items {
			item := testItemsToCreate[item]

			t.Run(fmt.Sprintf("%s_%s", tc.name, item.Attrs["category"]), func(t *testing.T) {
				t.Parallel()

				// Generate unique identifier for this test run to avoid conflicts in parallel execution
				uniqueID := uuid.New().String()

				var itemUUID string

				createAttrs := map[string]any{
					"title":    addUniqueIDToTitle(item.Attrs["title"].(string), uniqueID),
					"category": item.Attrs["category"],
					"section":  sections.MapSections(tc.create.Sections),
				}

				updateAttrs := map[string]any{
					"title":    addUniqueIDToTitle(item.Attrs["title"].(string), uniqueID),
					"category": item.Attrs["category"],
					"section":  sections.MapSections(tc.update.Sections),
				}

				// Build check functions for create step
				createChecks := []resource.TestCheckFunc{
					logStep(t, "CREATE"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
				}
				createChecks = append(createChecks, checks.BuildItemChecks("onepassword_item.test_item", createAttrs)...)

				// Build check functions for update step
				updateChecks := []resource.TestCheckFunc{
					logStep(t, "UPDATE"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
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

func TestAccItemResourceTags(t *testing.T) {
	t.Parallel()

	// Generate unique identifier for this test run to avoid conflicts in parallel execution
	uniqueID := uuid.New().String()

	item := testItemsToCreate[model.Login]

	testCases := []struct {
		name string
		tags []string
	}{
		{"CREATE_ITEM_WITH_2_TAGS", []string{"firstTestTag", "secondTestTag"}},
		{"ADD_3RD_TAG", []string{"firstTestTag", "secondTestTag", "thirdTestTag"}},
		{"REMOVE_2_TAGS", []string{"firstTestTag"}},
	}

	testVaultID := vault.GetTestVaultID(t)

	var testSteps []resource.TestStep

	for i, step := range testCases {
		attrs := maps.Clone(item.Attrs)
		attrs["title"] = addUniqueIDToTitle(attrs["title"].(string), uniqueID)
		attrs["tags"] = step.tags

		var itemUUID string
		testChecks := []resource.TestCheckFunc{}

		// Capture UUID and register cleanup
		if i == 0 {
			testChecks = append(testChecks,
				uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
				cleanup.RegisterItem(t, &itemUUID, testVaultID),
			)
		}

		testChecks = append(testChecks, logStep(t, step.name))
		testChecks = append(testChecks, checks.BuildItemChecks("onepassword_item.test_item", attrs)...)

		testSteps = append(testSteps, resource.TestStep{
			Config: tfconfig.CreateConfigBuilder()(
				tfconfig.ProviderConfig(),
				tfconfig.ItemResourceConfig(testVaultID, attrs),
			),
			Check: resource.ComposeAggregateTestCheckFunc(testChecks...),
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testSteps,
	})
}

func TestAccRecreateNonExistingItem(t *testing.T) {
	t.Parallel()

	// Generate unique identifier for this test run to avoid conflicts in parallel execution
	uniqueID := uuid.New().String()

	item := testItemsToCreate[model.Login]
	testVaultID := vault.GetTestVaultID(t)

	// Create a copy of item attributes and update title with unique ID
	createAttrs := maps.Clone(item.Attrs)
	createAttrs["title"] = addUniqueIDToTitle(createAttrs["title"].(string), uniqueID)

	var itemUUID string

	// Build check functions for create step
	createChecks := []resource.TestCheckFunc{
		logStep(t, "CREATE"),
		uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
		cleanup.RegisterItem(t, &itemUUID, testVaultID),
	}
	bcCreate := checks.BuildItemChecks("onepassword_item.test_item", createAttrs)
	createChecks = append(createChecks, bcCreate...)

	// Build check function to manually delete the item after creation
	deleteItemCheck := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			t.Log("MANUALLY_DELETE_ITEM")
			ctx := context.Background()

			client, err := client.CreateTestClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			itemToDelete := &model.Item{
				ID:      itemUUID,
				VaultID: testVaultID,
			}
			err = client.DeleteItem(ctx, itemToDelete, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to delete item: %w", err)
			}

			t.Logf("Successfully deleted item %s from vault %s", itemUUID, testVaultID)
			return nil
		}
	}

	// Build check functions for recreate step - verify the item was recreated
	recreateChecks := []resource.TestCheckFunc{
		logStep(t, "RECREATE"),
		uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
		cleanup.RegisterItem(t, &itemUUID, testVaultID),
	}
	bcRecreate := checks.BuildItemChecks("onepassword_item.test_item", createAttrs)
	recreateChecks = append(recreateChecks, bcRecreate...)

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
			// Manually delete the item outside of Terraform
			// After this step, Terraform will refresh and detect the item is missing,
			// so it will plan to recreate it. We expect a non-empty plan.
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check:              deleteItemCheck(),
				ExpectNonEmptyPlan: true,
			},
			// Run Terraform again - it should detect the item is missing and recreate it
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(recreateChecks...),
			},
		},
	})
}

func TestAccItemResource_DetectManualChanges(t *testing.T) {
	t.Parallel()

	// Generate unique identifier for this test run to avoid conflicts in parallel execution
	uniqueID := uuid.New().String()
	var itemUUID string
	testVaultID := vault.GetTestVaultID(t)

	initialAttrs := maps.Clone(testItemsToCreate[model.Login].Attrs)

	initialAttrs["title"] = addUniqueIDToTitle(initialAttrs["title"].(string), uniqueID)
	initialAttrs["section"] = sections.MapSections([]sections.TestSection{
		{
			Label: "Original Section",
			Fields: []sections.TestField{
				{Label: "Original Field 1", Value: "original value 1", Type: "STRING"},
				{Label: "Original Field 2", Value: "original value 2", Type: "EMAIL"},
			},
		},
	})

	updatedAttrs := maps.Clone(testItemsUpdatedAttrs[model.Login])
	updatedAttrs["title"] = initialAttrs["title"]
	updatedAttrs["section"] = sections.MapSections([]sections.TestSection{
		{
			Label: "Updated Section",
			Fields: []sections.TestField{
				{Label: "New Field", Value: "new value", Type: "URL"},
			},
		},
	})

	removedAttrs := map[string]any{
		"title":      initialAttrs["title"],
		"category":   "login",
		"username":   "",
		"note_value": "",
		"url":        []string{},
		"tags":       []string{},
		"section":    []map[string]any{},
	}

	// Build check functions for create step
	createChecks := []resource.TestCheckFunc{
		logStep(t, "CREATE"),
		uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
		cleanup.RegisterItem(t, &itemUUID, testVaultID),
	}
	bcCreate := checks.BuildItemChecks("onepassword_item.test_item", initialAttrs)
	createChecks = append(createChecks, bcCreate...)

	// Build check function to manually update the item after creation
	updateItemOutsideTerraform := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			t.Log("MANUALLY_UPDATE_ITEM")

			ctx := context.Background()
			client, err := client.CreateTestClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			currentItem := &model.Item{
				ID:       itemUUID,
				VaultID:  testVaultID,
				Category: model.Login,
			}

			updatedItem := attributes.BuildUpdatedItemAttrs(currentItem, updatedAttrs)
			_, err = client.UpdateItem(ctx, updatedItem, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to update item: %w", err)
			}

			return nil
		}
	}

	// Build check function to manually remove all fields
	removeFieldsOutsideTerraform := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			t.Log("MANUALLY_REMOVE_ALL_FIELDS")
			ctx := context.Background()

			client, err := client.CreateTestClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			strippedItem := &model.Item{
				ID:       itemUUID,
				Title:    removedAttrs["title"].(string),
				VaultID:  testVaultID,
				Category: model.Login,
				Tags:     []string{},
				URLs: []model.ItemURL{
					{URL: "", Primary: true},
				},
				Sections: []model.ItemSection{},
				Fields:   []model.ItemField{},
			}

			_, err = client.UpdateItem(ctx, strippedItem, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to remove fields: %w", err)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create new item
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(createChecks...),
			},
			// Manually update the item outside of Terraform
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
				),
				Check:              updateItemOutsideTerraform(),
				ExpectNonEmptyPlan: true,
			},
			// Verify manual updates via import
			{
				ResourceName:      "onepassword_item.test_item",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, initialAttrs["title"]),
				ImportStateVerify: false,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					t.Log("VERIFY_MANUAL_UPDATES")
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					state := states[0]
					expectedAttrs := attributes.BuildImportAttrs(updatedAttrs)

					for key, expected := range expectedAttrs {
						if actual := state.Attributes[key]; actual != expected {
							return fmt.Errorf("%s: expected %v, got %v", key, expected, actual)
						}
					}
					return nil
				},
			},
			// Manually remove all fields
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
				),
				Check:              removeFieldsOutsideTerraform(),
				ExpectNonEmptyPlan: true,
			},
			// Verify fields were removed
			{
				ResourceName:      "onepassword_item.test_item",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, initialAttrs["title"]),
				ImportStateVerify: false,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					t.Log("VERIFY_FIELDS_REMOVED")
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					state := states[0]

					// Check that fields are empty/removed
					checks := map[string]any{
						"title":     initialAttrs["title"],
						"category":  "login",
						"username":  "",
						"url":       "",
						"tags":      "",
						"section.#": "0",
					}
					for key, expected := range checks {
						if actual := state.Attributes[key]; actual != expected {
							return fmt.Errorf("%s: expected %q, got %q", key, expected, actual)
						}
					}

					return nil
				},
			},
		},
	})
}

func TestAccItemResourcePasswordGenerationForAllCategories(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)

	// Test all three categories that support password generation
	categories := []struct {
		name     string
		category model.ItemCategory
		attrs    map[string]any
	}{
		{
			name:     "Login",
			category: model.Login,
			attrs: map[string]any{
				"title":    "Test Login Password Generation",
				"category": "login",
				"username": "testuser@example.com",
				"url":      "https://example.com",
				"password_recipe": password.BuildPasswordRecipeMap(password.PasswordRecipe{
					Length:  20,
					Symbols: true,
					Digits:  true,
				}),
			},
		},
		{
			name:     "Password",
			category: model.Password,
			attrs: map[string]any{
				"title":    "Test Password Category Generation",
				"category": "password",
				"password_recipe": password.BuildPasswordRecipeMap(password.PasswordRecipe{
					Length:  20,
					Symbols: true,
					Digits:  true,
				}),
			},
		},
		{
			name:     "Database",
			category: model.Database,
			attrs: map[string]any{
				"title":    "Test Database Password Generation",
				"category": "database",
				"database": "testdatabase",
				"username": "testusername",
				"password_recipe": password.BuildPasswordRecipeMap(password.PasswordRecipe{
					Length:  20,
					Symbols: true,
					Digits:  true,
				}),
			},
		},
	}

	recipe := password.PasswordRecipe{
		Length:  20,
		Symbols: true,
		Digits:  true,
	}

	for _, tc := range categories {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uniqueID := uuid.New().String()
			var itemUUID string

			attrs := maps.Clone(tc.attrs)
			attrs["title"] = addUniqueIDToTitle(attrs["title"].(string), uniqueID)

			// Build checks to verify password was generated
			checks := []resource.TestCheckFunc{
				uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
				cleanup.RegisterItem(t, &itemUUID, testVaultID),
			}
			checks = append(checks, password.BuildPasswordRecipeChecks("onepassword_item.test_item", recipe)...)
			checks = append(checks, resource.TestCheckResourceAttrSet("onepassword_item.test_item", "password"))

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, attrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(checks...),
					},
				},
			})
		})
	}
}

func TestAccItemResourceEmptyStringPreservation(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	var itemUUID string

	attrs := map[string]any{
		"title":      "",
		"category":   "database",
		"username":   "",
		"url":        "",
		"hostname":   "",
		"database":   "",
		"port":       "",
		"note_value": "",
		"section": []map[string]any{
			{
				"label": "",
				"field": []map[string]any{
					{
						"label": "test_field",
						"value": "",
						"type":  "STRING",
					},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "url", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "hostname", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "database", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "port", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", ""),
				),
			},
		},
	})
}

func TestAccItemResourceNullVsEmptyString(t *testing.T) {
	t.Parallel()

	var itemUUID string
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()

	attrsWithoutFields := map[string]any{
		"title":    addUniqueIDToTitle("Test Null vs Empty", uniqueID),
		"category": "database",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrsWithoutFields),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "username"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "url"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "hostname"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "database"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "port"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value"),
				),
			},
		},
	})
}

func TestAccItemResourceClearFieldsToEmptyString(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Clear Fields", uniqueID)
	var itemUUID string

	attrsWithValues := map[string]any{
		"title":      title,
		"category":   "database",
		"username":   "testuser",
		"hostname":   "db.example.com",
		"database":   "mydb",
		"port":       "3306",
		"note_value": "test_note",
		"section": []map[string]any{
			{
				"label": "test_section",
				"field": []map[string]any{
					{
						"label": "test_field",
						"value": "test_value",
						"type":  "STRING",
					},
				},
			},
		},
	}

	attrsCleared := map[string]any{
		"title":      title,
		"category":   "database",
		"username":   "",
		"hostname":   "",
		"database":   "",
		"port":       "",
		"note_value": "",
		"section": []map[string]any{
			{
				"label": "",
				"field": []map[string]any{
					{
						"label": "",
						"value": "",
						"type":  "STRING",
					},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrsWithValues),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "testuser"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "hostname", "db.example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "database", "mydb"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "port", "3306"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", "test_section"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.label", "test_field"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value", "test_note"),
				),
			},
			// Clear all fields
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrsCleared),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "hostname", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "database", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "port", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.label", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value", ""),
				),
			},
		},
	})
}

// TestAccItemResourcePasswordWriteOnly tests the password_wo (write-only) functionality
func TestAccItemResourcePasswordWriteOnly(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Password Write-Only", uniqueID)

	var itemUUID string

	// Step 1: Create item with password_wo
	createAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "testuser@example.com",
		"password_wo":         "initial-password-123",
		"password_wo_version": 1,
	}

	// Step 2: Update password by incrementing version
	updatePasswordAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "testuser@example.com",
		"password_wo":         "updated-password-456",
		"password_wo_version": 2,
	}

	// Step 3: Update other fields without changing password (version unchanged)
	updateOtherFieldsAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "updateduser@example.com",
		"url":                 "https://example.com",
		"password_wo":         "updated-password-456", // Same password, but won't be in plan
		"password_wo_version": 2,                      // Same version - password should be preserved
	}

	// Step 4: Add section while preserving password
	updateWithSectionAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "updateduser@example.com",
		"url":                 "https://example.com",
		"password_wo":         "updated-password-456",
		"password_wo_version": 2, // Same version - password should be preserved
		"section": sections.MapSections([]sections.TestSection{
			{
				Label: "Test Section",
				Fields: []sections.TestField{
					{Label: "Test Field", Value: "test-value", Type: "STRING"},
				},
			},
		}),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with password_wo
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_PASSWORD_WO"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", title),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "login"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "testuser@example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "1"),
					// Verify password_wo is not in state (write-only)
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "password_wo"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "password"),
					// Verify password was set in 1Password by checking via client
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "initial-password-123"),
				),
			},
			// Step 2: Update password by incrementing version
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updatePasswordAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "UPDATE_PASSWORD_VERSION_INCREMENT"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "password_wo"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "password"),
					// Verify password was updated in 1Password
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "updated-password-456"),
				),
			},
			// Step 3: Update other fields without changing password (version unchanged)
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateOtherFieldsAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "UPDATE_OTHER_FIELDS_PRESERVE_PASSWORD"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "updateduser@example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "url", "https://example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "2"),
					// Verify password was preserved (not changed)
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "updated-password-456"),
				),
			},
			// Step 4: Add section while preserving password
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateWithSectionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "ADD_SECTION_PRESERVE_PASSWORD"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.#", "1"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", "Test Section"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.#", "1"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.label", "Test Field"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.value", "test-value"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "2"),
					// Verify password was preserved when adding section
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "updated-password-456"),
				),
			},
		},
	})
}

// TestAccItemResourcePasswordWriteOnlyVersionDecrement tests that password is not updated when version is decremented
func TestAccItemResourcePasswordWriteOnlyVersionDecrement(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Password WO Version Decrement", uniqueID)

	var itemUUID string

	createAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "testuser@example.com",
		"password_wo":         "initial-password-123",
		"password_wo_version": 2,
	}

	// Try to decrement version (should not update password)
	decrementVersionAttrs := map[string]any{
		"title":               title,
		"category":            "login",
		"username":            "testuser@example.com",
		"password_wo":         "should-not-be-used",
		"password_wo_version": 1, // Decremented - password should not be updated
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with version 2
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_VERSION_2"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "2"),
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "initial-password-123"),
				),
			},
			// Try to decrement version - password should not be updated
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, decrementVersionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "DECREMENT_VERSION"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password_wo_version", "1"),
					// Password should still be the original one (not updated)
					verifyPasswordIn1Password(t, testVaultID, &itemUUID, "initial-password-123"),
				),
			},
		},
	})
}

// TestAccItemResourceNoteValueWriteOnly tests the note_value_wo (write-only) functionality
func TestAccItemResourceNoteValueWriteOnly(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Note Value Write-Only", uniqueID)

	var itemUUID string

	// Step 1: Create item with note_value_wo
	createAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"note_value_wo":         "initial-note-value-123",
		"note_value_wo_version": 1,
	}

	// Step 2: Update note_value by incrementing version
	updateNoteValueAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"note_value_wo":         "updated-note-value-456",
		"note_value_wo_version": 2,
	}

	// Step 3: Update other fields without changing note_value (version unchanged)
	updateOtherFieldsAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"tags":                  []string{"tag1", "tag2"},
		"note_value_wo":         "updated-note-value-456", // Same note_value, but won't be in plan
		"note_value_wo_version": 2,                        // Same version - note_value should be preserved
	}

	// Step 4: Add section while preserving note_value
	updateWithSectionAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"tags":                  []string{"tag1", "tag2"},
		"note_value_wo":         "updated-note-value-456",
		"note_value_wo_version": 2, // Same version - note_value should be preserved
		"section": sections.MapSections([]sections.TestSection{
			{
				Label: "Test Section",
				Fields: []sections.TestField{
					{Label: "Test Field", Value: "test-value", Type: "STRING"},
				},
			},
		}),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with note_value_wo
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_NOTE_VALUE_WO"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", title),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "secure_note"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "1"),
					// Verify note_value_wo is not in state (write-only)
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value_wo"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value"),
					// Verify note_value was set in 1Password by checking via client
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "initial-note-value-123"),
				),
			},
			// Step 2: Update note_value by incrementing version
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateNoteValueAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "UPDATE_NOTE_VALUE_VERSION_INCREMENT"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "2"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value_wo"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value"),
					// Verify note_value was updated in 1Password
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "updated-note-value-456"),
				),
			},
			// Step 3: Update other fields without changing note_value (version unchanged)
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateOtherFieldsAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "UPDATE_OTHER_FIELDS_PRESERVE_NOTE_VALUE"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "tags.#", "2"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "2"),
					// Verify note_value was preserved (not changed)
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "updated-note-value-456"),
				),
			},
			// Step 4: Add section while preserving note_value
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateWithSectionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "ADD_SECTION_PRESERVE_NOTE_VALUE"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.#", "1"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", "Test Section"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.#", "1"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.label", "Test Field"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.value", "test-value"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "2"),
					// Verify note_value was preserved when adding section
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "updated-note-value-456"),
				),
			},
		},
	})
}

// TestAccItemResourceNoteValueWriteOnlyVersionDecrement tests that note_value is not updated when version is decremented
func TestAccItemResourceNoteValueWriteOnlyVersionDecrement(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Note Value WO Version Decrement", uniqueID)

	var itemUUID string

	createAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"note_value_wo":         "initial-note-value-123",
		"note_value_wo_version": 2,
	}

	// Try to decrement version (should not update note_value)
	decrementVersionAttrs := map[string]any{
		"title":                 title,
		"category":              "secure_note",
		"note_value_wo":         "should-not-be-used",
		"note_value_wo_version": 1, // Decremented - note_value should not be updated
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with version 2
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_VERSION_2"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "2"),
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "initial-note-value-123"),
				),
			},
			// Try to decrement version - note_value should not be updated
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, decrementVersionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "DECREMENT_VERSION"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value_wo_version", "1"),
					// Note value should still be the original one (not updated)
					verifyNoteValueIn1Password(t, testVaultID, &itemUUID, "initial-note-value-123"),
				),
			},
		},
	})
}

// verifyFieldValueIn1Password verifies that a field value in 1Password matches the expected value
func verifyFieldValueIn1Password(t *testing.T, vaultID string, itemUUID *string, fieldPurpose model.ItemFieldPurpose, fieldName string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client, err := client.CreateTestClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		item, err := client.GetItem(ctx, *itemUUID, vaultID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}

		// Find field with the specified purpose
		for _, f := range item.Fields {
			if f.Purpose == fieldPurpose {
				if f.Value != expectedValue {
					return fmt.Errorf("%s mismatch: expected %q, got %q", fieldName, expectedValue, f.Value)
				}
				t.Logf("%s verified in 1Password: %q", fieldName, f.Value)
				return nil
			}
		}

		// If field not found and expected value is empty, that's OK
		if expectedValue == "" {
			t.Logf("%s field not found in 1Password (as expected)", fieldName)
			return nil
		}

		return fmt.Errorf("%s field not found in item", fieldName)
	}
}

// verifyNoteValueIn1Password verifies that the note_value in 1Password matches the expected value
func verifyNoteValueIn1Password(t *testing.T, vaultID string, itemUUID *string, expectedNoteValue string) resource.TestCheckFunc {
	return verifyFieldValueIn1Password(t, vaultID, itemUUID, model.FieldPurposeNotes, "note_value", expectedNoteValue)
}

// verifyPasswordIn1Password verifies that the password in 1Password matches the expected value
func verifyPasswordIn1Password(t *testing.T, vaultID string, itemUUID *string, expectedPassword string) resource.TestCheckFunc {
	return verifyFieldValueIn1Password(t, vaultID, itemUUID, model.FieldPurposePassword, "password", expectedPassword)
}

// addUniqueIDToTitle appends a UUID to the title to avoid conflicts in parallel test execution
func addUniqueIDToTitle(title string, uniqueID string) string {
	return fmt.Sprintf("%s-%s", title, uniqueID)
}

// logStep logs the current test step for easier test debugging
func logStep(t *testing.T, step string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Log(step)
		return nil
	}
}

func TestAccItemResourceSectionMap_BasicCRUD(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap CRUD", uniqueID)
	var itemUUID string

	// Step 1: Create with single section and multiple field types
	createSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"api_key": {
					Type:  "STRING",
					Value: "initial-api-key",
				},
				"api_secret": {
					Type:  "CONCEALED",
					Value: "initial-secret",
				},
			},
		},
	})

	createAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"username":    "testuser@example.com",
		"section_map": createSectionMap,
	}

	// Step 2: Update field values
	updateValuesSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"api_key": {
					Type:  "STRING",
					Value: "updated-api-key",
				},
				"api_secret": {
					Type:  "CONCEALED",
					Value: "updated-secret",
				},
			},
		},
	})

	updateValuesAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"username":    "testuser@example.com",
		"section_map": updateValuesSectionMap,
	}

	// Step 3: Add new field to existing section
	addFieldSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"api_key": {
					Type:  "STRING",
					Value: "updated-api-key",
				},
				"api_secret": {
					Type:  "CONCEALED",
					Value: "updated-secret",
				},
				"environment": {
					Type:  "STRING",
					Value: "production",
				},
			},
		},
	})

	addFieldAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"username":    "testuser@example.com",
		"section_map": addFieldSectionMap,
	}

	// Step 4: Add new section
	addSectionSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"api_key": {
					Type:  "STRING",
					Value: "updated-api-key",
				},
				"api_secret": {
					Type:  "CONCEALED",
					Value: "updated-secret",
				},
				"environment": {
					Type:  "STRING",
					Value: "production",
				},
			},
		},
		"metadata": {
			FieldMap: map[string]sections.TestSectionMapField{
				"created_by": {
					Type:  "STRING",
					Value: "terraform",
				},
			},
		},
	})

	addSectionAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"username":    "testuser@example.com",
		"section_map": addSectionSectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_SECTION_MAP"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", title),
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "credentials"),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "credentials", "api_key"),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "credentials", "api_secret"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_key", "initial-api-key"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_secret", "initial-secret"),
				),
			},
			// Step 2: Read/Refresh - verify no changes
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "READ_REFRESH_NO_CHANGES"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
				),
			},
			// Step 3: Update field values
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, updateValuesAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "UPDATE_FIELD_VALUES"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_key", "updated-api-key"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_secret", "updated-secret"),
				),
			},
			// Step 4: Add new field to existing section
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, addFieldAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "ADD_FIELD_TO_SECTION"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "credentials", "environment"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "environment", "production"),
				),
			},
			// Step 5: Add new section
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, addSectionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "ADD_NEW_SECTION"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "metadata"),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "metadata", "created_by"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "metadata", "created_by", "terraform"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_FieldTypes tests all supported field types in section_map:
func TestAccItemResourceSectionMap_FieldTypes(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Field Types", uniqueID)
	var itemUUID string

	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"all_types": {
			FieldMap: map[string]sections.TestSectionMapField{
				"string_field": {
					Type:  "STRING",
					Value: "plain text value",
				},
				"concealed_field": {
					Type:  "CONCEALED",
					Value: "secret-value-123",
				},
				"email_field": {
					Type:  "EMAIL",
					Value: "test@example.com",
				},
				"url_field": {
					Type:  "URL",
					Value: "https://example.com",
				},
				"date_field": {
					Type:  "DATE",
					Value: "2025-01-06",
				},
				"month_year_field": {
					Type:  "MONTH_YEAR",
					Value: "202501",
				},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": sectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_ALL_FIELD_TYPES"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					// Verify all field types
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.string_field.type", "STRING"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.string_field.value", "plain text value"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.concealed_field.type", "CONCEALED"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.concealed_field.value", "secret-value-123"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.email_field.type", "EMAIL"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.email_field.value", "test@example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.url_field.type", "URL"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.url_field.value", "https://example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.date_field.type", "DATE"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.date_field.value", "2025-01-06"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.month_year_field.type", "MONTH_YEAR"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.all_types.field_map.month_year_field.value", "202501"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_PasswordRecipe tests password generation in section_map fields:
func TestAccItemResourceSectionMap_PasswordRecipe(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		length  int
		digits  bool
		symbols bool
	}{
		{name: "DefaultRecipe", length: 20, digits: true, symbols: true},
		{name: "LongPassword", length: 64, digits: true, symbols: true},
		{name: "NoDigits", length: 20, digits: false, symbols: true},
		{name: "NoSymbols", length: 20, digits: true, symbols: false},
		{name: "LettersOnly", length: 20, digits: false, symbols: false},
	}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uniqueID := uuid.New().String()
			title := addUniqueIDToTitle("Test SectionMap Password Recipe", uniqueID)
			var itemUUID string

			recipe := map[string]any{
				"length":  tc.length,
				"digits":  tc.digits,
				"symbols": tc.symbols,
			}

			sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
				"credentials": {
					FieldMap: map[string]sections.TestSectionMapField{
						"generated_password": {
							Type:           "CONCEALED",
							PasswordRecipe: &recipe,
						},
					},
				},
			})

			attrs := map[string]any{
				"title":       title,
				"category":    "login",
				"section_map": sectionMap,
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, attrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							logStep(t, fmt.Sprintf("CREATE_WITH_PASSWORD_RECIPE_%s", tc.name)),
							uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
							cleanup.RegisterItem(t, &itemUUID, testVaultID),
							// Verify password was generated (value is set)
							resource.TestCheckResourceAttrSet("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.value"),
							// Verify recipe attributes are preserved
							resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.length", fmt.Sprintf("%d", tc.length)),
							resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.digits", fmt.Sprintf("%t", tc.digits)),
							resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.symbols", fmt.Sprintf("%t", tc.symbols)),
						),
					},
				},
			})
		})
	}
}

// TestAccItemResourceSectionMap_ValueAndPasswordRecipeConflict tests that value and password_recipe
// cannot be specified together in section_map fields:
func TestAccItemResourceSectionMap_ValueAndPasswordRecipeConflict(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Value Recipe Conflict", uniqueID)

	recipe := map[string]any{
		"length":  20,
		"digits":  true,
		"symbols": true,
	}

	// Create section_map with both value and password_recipe set - this should fail validation
	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"conflicting_field": {
					Type:           "CONCEALED",
					Value:          "my-explicit-value",
					PasswordRecipe: &recipe,
				},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": sectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func TestAccItemResourceSectionMap_MultipleSections(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Multiple Sections", uniqueID)
	var itemUUID string

	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"database": {
			FieldMap: map[string]sections.TestSectionMapField{
				"host": {Type: "STRING", Value: "db.example.com"},
				"port": {Type: "STRING", Value: "5432"},
				"name": {Type: "STRING", Value: "mydb"},
			},
		},
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"username": {Type: "STRING", Value: "admin"},
				"password": {Type: "CONCEALED", Value: "secret123"},
			},
		},
		"metadata": {
			FieldMap: map[string]sections.TestSectionMapField{
				"environment": {Type: "STRING", Value: "production"},
				"owner":       {Type: "EMAIL", Value: "team@example.com"},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "database",
		"section_map": sectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_MULTIPLE_SECTIONS"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					// Verify database section
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "database"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "database", "host", "db.example.com"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "database", "port", "5432"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "database", "name", "mydb"),
					// Verify credentials section
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "credentials"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "username", "admin"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "password", "secret123"),
					// Verify metadata section
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "metadata"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "metadata", "environment", "production"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "metadata", "owner", "team@example.com"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_RemoveFieldAndSection tests removal operations
func TestAccItemResourceSectionMap_RemoveFieldAndSection(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Remove", uniqueID)
	var itemUUID string

	// Step 1: Create with multiple sections and fields
	initialSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"section1": {
			FieldMap: map[string]sections.TestSectionMapField{
				"field1": {Type: "STRING", Value: "value1"},
				"field2": {Type: "STRING", Value: "value2"},
			},
		},
		"section2": {
			FieldMap: map[string]sections.TestSectionMapField{
				"field3": {Type: "STRING", Value: "value3"},
			},
		},
	})

	initialAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": initialSectionMap,
	}

	// Step 2: Remove field2 from section1
	removeFieldSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"section1": {
			FieldMap: map[string]sections.TestSectionMapField{
				"field1": {Type: "STRING", Value: "value1"},
			},
		},
		"section2": {
			FieldMap: map[string]sections.TestSectionMapField{
				"field3": {Type: "STRING", Value: "value3"},
			},
		},
	})

	removeFieldAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": removeFieldSectionMap,
	}

	// Step 3: Remove section2 entirely
	removeSectionSectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"section1": {
			FieldMap: map[string]sections.TestSectionMapField{
				"field1": {Type: "STRING", Value: "value1"},
			},
		},
	})

	removeSectionAttrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": removeSectionSectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_INITIAL"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section1", "field1", "value1"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section1", "field2", "value2"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section2", "field3", "value3"),
				),
			},
			// Step 2: Remove field
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, removeFieldAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "REMOVE_FIELD"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section1", "field1", "value1"),
					// field2 should be gone - check no attribute
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "section_map.section1.field_map.field2.value"),
				),
			},
			// Step 3: Remove section
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, removeSectionAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "REMOVE_SECTION"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "section1"),
					// section2 should be gone
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "section_map.section2.id"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_WithPasswordRecipeAndOtherFields
func TestAccItemResourceSectionMap_WithPasswordRecipeAndOtherFields(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Mixed Fields", uniqueID)
	var itemUUID string

	recipe := map[string]any{
		"length":  30,
		"digits":  true,
		"symbols": true,
	}

	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"credentials": {
			FieldMap: map[string]sections.TestSectionMapField{
				"api_key": {
					Type:  "STRING",
					Value: "my-api-key-123",
				},
				"api_secret": {
					Type:  "CONCEALED",
					Value: "my-secret-value",
				},
				"generated_password": {
					Type:           "CONCEALED",
					PasswordRecipe: &recipe,
				},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "login",
		"username":    "testuser@example.com",
		"section_map": sectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_MIXED_FIELDS"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					// Verify regular fields
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_key", "my-api-key-123"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "credentials", "api_secret", "my-secret-value"),
					// Verify generated password field
					resource.TestCheckResourceAttrSet("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.value"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.length", "30"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.digits", "true"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section_map.credentials.field_map.generated_password.password_recipe.symbols", "true"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_EmptyValues tests handling of empty values
func TestAccItemResourceSectionMap_EmptyValues(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SectionMap Empty Values", uniqueID)
	var itemUUID string

	// Create with empty value field
	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"section1": {
			FieldMap: map[string]sections.TestSectionMapField{
				"empty_field": {
					Type:  "STRING",
					Value: "", // Empty value
				},
				"filled_field": {
					Type:  "STRING",
					Value: "has-value",
				},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": sectionMap,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_EMPTY_VALUES"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "section1"),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "section1", "empty_field"),
					checks.BuildSectionMapFieldIDSetCheck("onepassword_item.test_item", "section1", "filled_field"),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section1", "filled_field", "has-value"),
				),
			},
		},
	})
}

// TestAccItemResourceSectionMap_AllCategories tests section_map with different item categories
func TestAccItemResourceSectionMap_AllCategories(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		category string
		attrs    map[string]any
	}{
		{
			name:     "Login",
			category: "login",
			attrs: map[string]any{
				"username": "testuser@example.com",
			},
		},
		{
			name:     "Password",
			category: "password",
			attrs:    map[string]any{},
		},
		{
			name:     "Database",
			category: "database",
			attrs: map[string]any{
				"database": "testdb",
				"hostname": "localhost",
			},
		},
		{
			name:     "SecureNote",
			category: "secure_note",
			attrs: map[string]any{
				"note_value": "This is a secure note",
			},
		},
	}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uniqueID := uuid.New().String()
			title := addUniqueIDToTitle(fmt.Sprintf("Test SectionMap %s", tc.name), uniqueID)
			var itemUUID string

			sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
				"custom_section": {
					FieldMap: map[string]sections.TestSectionMapField{
						"custom_field": {
							Type:  "STRING",
							Value: "custom-value",
						},
					},
				},
			})

			attrs := map[string]any{
				"title":       title,
				"category":    tc.category,
				"section_map": sectionMap,
			}

			// Merge additional attrs
			for k, v := range tc.attrs {
				attrs[k] = v
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, attrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							logStep(t, fmt.Sprintf("CREATE_%s_WITH_SECTION_MAP", tc.name)),
							uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
							cleanup.RegisterItem(t, &itemUUID, testVaultID),
							resource.TestCheckResourceAttr("onepassword_item.test_item", "category", tc.category),
							checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "custom_section"),
							checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "custom_section", "custom_field", "custom-value"),
						),
					},
				},
			})
		})
	}
}

// TestAccItemResourceSectionMap_DuplicateKeys tests behavior when duplicate map keys are used
func TestAccItemResourceSectionMap_DuplicateKeys(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)

	t.Run("DuplicateSectionKeys_LastWins", func(t *testing.T) {
		t.Parallel()

		uniqueID := uuid.New().String()
		var itemUUID string

		config := fmt.Sprintf(`
provider "onepassword" {}

resource "onepassword_item" "test_item" {
  vault    = "%s"
  title    = "Test Duplicate Section Keys-%s"
  category = "login"

  section_map = {
    "duplicate_section" = {
      field_map = {
        "field1" = {
          type  = "STRING"
          value = "first_value_should_be_overwritten"
        }
      }
    }
    "duplicate_section" = {
      field_map = {
        "field2" = {
          type  = "STRING"
          value = "second_value_wins"
        }
      }
    }
  }
}
`, testVaultID, uniqueID)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						logStep(t, "CREATE_DUPLICATE_SECTION_KEYS"),
						uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
						cleanup.RegisterItem(t, &itemUUID, testVaultID),
						// Only one section should exist (the second definition)
						checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "duplicate_section"),
						// field2 from second definition should exist
						checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "duplicate_section", "field2", "second_value_wins"),
						// field1 from first definition should NOT exist (was overwritten)
						resource.TestCheckNoResourceAttr("onepassword_item.test_item", "section_map.duplicate_section.field_map.field1"),
					),
				},
			},
		})
	})

	t.Run("DuplicateFieldKeys_LastWins", func(t *testing.T) {
		t.Parallel()

		uniqueID := uuid.New().String()
		var itemUUID string

		config := fmt.Sprintf(`
provider "onepassword" {}

resource "onepassword_item" "test_item" {
  vault    = "%s"
  title    = "Test Duplicate Field Keys-%s"
  category = "login"

  section_map = {
    "my_section" = {
      field_map = {
        "duplicate_field" = {
          type  = "STRING"
          value = "first_value_should_be_overwritten"
        }
        "duplicate_field" = {
          type  = "STRING"
          value = "second_value_wins"
        }
      }
    }
  }
}
`, testVaultID, uniqueID)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						logStep(t, "CREATE_DUPLICATE_FIELD_KEYS"),
						uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
						cleanup.RegisterItem(t, &itemUUID, testVaultID),
						// Section should exist
						checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "my_section"),
						// Only one field should exist with the second (last) value
						checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "my_section", "duplicate_field", "second_value_wins"),
					),
				},
			},
		})
	})

	t.Run("SameFieldLabelInDifferentSections_Success", func(t *testing.T) {
		t.Parallel()

		// Same field label in different sections is valid - each section has its own field_map
		uniqueID := uuid.New().String()
		title := addUniqueIDToTitle("Test Same Field Different Sections", uniqueID)
		var itemUUID string

		sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
			"section_a": {
				FieldMap: map[string]sections.TestSectionMapField{
					"common_field": {Type: "STRING", Value: "value in A"},
				},
			},
			"section_b": {
				FieldMap: map[string]sections.TestSectionMapField{
					"common_field": {Type: "STRING", Value: "value in B"},
				},
			},
		})

		attrs := map[string]any{
			"title":       title,
			"category":    "login",
			"section_map": sectionMap,
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfconfig.CreateConfigBuilder()(
						tfconfig.ProviderConfig(),
						tfconfig.ItemResourceConfig(testVaultID, attrs),
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						logStep(t, "CREATE_SAME_FIELD_DIFFERENT_SECTIONS"),
						uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
						cleanup.RegisterItem(t, &itemUUID, testVaultID),
						resource.TestCheckResourceAttr("onepassword_item.test_item", "title", title),
						checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "section_a"),
						checks.BuildSectionMapIDSetCheck("onepassword_item.test_item", "section_b"),
						checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section_a", "common_field", "value in A"),
						checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "section_b", "common_field", "value in B"),
					),
				},
			},
		})
	})
}

// TestAccItemResourceSectionMap_ConfigRestoresDriftedValue tests that when a field value
// is changed directly in 1Password (outside Terraform), running terraform apply
// restores the value from the config.
func TestAccItemResourceSectionMap_ConfigRestoresDriftedValue(t *testing.T) {
	t.Parallel()

	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Config Restores Drift", uniqueID)
	var itemUUID string
	var sectionID string

	configValue := "config-controlled-value"
	driftedValue := "manually-changed-value"

	// Create section map with the config value
	sectionMap := sections.BuildSectionMap(map[string]sections.TestSectionMapEntry{
		"my_section": {
			FieldMap: map[string]sections.TestSectionMapField{
				"my_field": {
					Type:  "STRING",
					Value: configValue,
				},
			},
		},
	})

	attrs := map[string]any{
		"title":       title,
		"category":    "login",
		"section_map": sectionMap,
	}

	captureSectionID := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["onepassword_item.test_item"]
			if !ok {
				return fmt.Errorf("resource not found in state")
			}

			sectionID = rs.Primary.Attributes["section_map.my_section.id"]
			if sectionID == "" {
				return fmt.Errorf("section_map.my_section.id not found in state")
			}

			return nil
		}
	}

	modifyFieldIn1Password := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			t.Log("MANUALLY_MODIFY_FIELD_VALUE_IN_1PASSWORD")
			ctx := context.Background()

			opClient, err := client.CreateTestClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			item, err := opClient.GetItem(ctx, itemUUID, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to get item: %w", err)
			}

			found := false
			for i, field := range item.Fields {
				if field.Label == "my_field" && field.SectionID == sectionID {
					item.Fields[i].Value = driftedValue
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("field 'my_field' not found in section %s", sectionID)
			}

			_, err = opClient.UpdateItem(ctx, item, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to update item: %w", err)
			}

			// Verify the change was persisted
			updatedItem, err := opClient.GetItem(ctx, itemUUID, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to verify updated item: %w", err)
			}

			for _, field := range updatedItem.Fields {
				if field.Label == "my_field" && field.SectionID == sectionID {
					if field.Value != driftedValue {
						return fmt.Errorf("field value was not updated: expected %q, got %q", driftedValue, field.Value)
					}
					t.Logf("Verified field value in 1Password is now: %q", field.Value)
					return nil
				}
			}

			return fmt.Errorf("field not found after update")
		}
	}

	// Helper to verify field value in 1Password matches expected value
	verifyFieldValueIn1Password := func(expectedValue string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			t.Logf("VERIFY_FIELD_VALUE_IN_1PASSWORD: expecting %q", expectedValue)
			ctx := context.Background()

			opClient, err := client.CreateTestClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			item, err := opClient.GetItem(ctx, itemUUID, testVaultID)
			if err != nil {
				return fmt.Errorf("failed to get item: %w", err)
			}

			for _, field := range item.Fields {
				if field.Label == "my_field" && field.SectionID == sectionID {
					if field.Value != expectedValue {
						return fmt.Errorf("field value mismatch in 1Password: expected %q, got %q", expectedValue, field.Value)
					}
					t.Logf("Field value in 1Password matches expected: %q", field.Value)
					return nil
				}
			}
			return fmt.Errorf("field 'my_field' not found in 1Password")
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create item with config value
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "CREATE_WITH_CONFIG_VALUE"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					captureSectionID(),
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "my_section", "my_field", configValue),
					verifyFieldValueIn1Password(configValue),
				),
			},
			// Step 2: Directly modify the field in 1Password (simulate drift)
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					modifyFieldIn1Password(),
				),
				// Expect non-empty plan because the value changed
				ExpectNonEmptyPlan: true,
			},
			// Step 3: Apply again with same config - should restore config value
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "APPLY_RESTORES_CONFIG_VALUE"),
					// Verify Terraform state has the config value
					checks.BuildSectionMapFieldValueCheck("onepassword_item.test_item", "my_section", "my_field", configValue),
					// Verify 1Password has the config value restored
					verifyFieldValueIn1Password(configValue),
				),
			},
			// Step 4: Verify no more changes needed
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, attrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					logStep(t, "VERIFY_NO_CHANGES"),
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_item", &itemUUID),
				),
			},
		},
	})
}
