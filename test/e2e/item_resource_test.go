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
					checks := password.BuildPasswordRecipeChecks("onepassword_item.test_item", tc.recipe)
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
				checks := password.BuildPasswordRecipeChecksForField("onepassword_item.test_item", "section.0.field.0", tc.recipe)
				testStep.Check = resource.ComposeAggregateTestCheckFunc(checks...)
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    []resource.TestStep{testStep},
			})
		})
	}
}

func TestAccItemResourceSectionsAndFields(t *testing.T) {
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

	for _, step := range testCases {
		attrs := maps.Clone(item.Attrs)
		attrs["title"] = addUniqueIDToTitle(attrs["title"].(string), uniqueID)
		attrs["tags"] = step.tags

		testChecks := []resource.TestCheckFunc{logStep(t, step.name)}
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

func TestAccItemResourceEmptyStringPreservation(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)

	attrs := map[string]any{
		"title":      "",
		"category":   "database",
		"username":   "",
		"url":        "",
		"hostname":   "",
		"database":   "",
		"port":       "",
		"note_value": "",
		"password":   "",
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
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "url", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "hostname", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "database", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "port", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", ""),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password", ""),
				),
			},
		},
	})
}

func TestAccItemResourceNullVsEmptyString(t *testing.T) {
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
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "username"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "url"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "hostname"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "database"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "port"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "note_value"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_item", "password"),
				),
			},
		},
	})
}

func TestAccItemResourceClearFieldsToEmptyString(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test Clear Fields", uniqueID)

	attrsWithValues := map[string]any{
		"title":      title,
		"category":   "database",
		"username":   "testuser",
		"hostname":   "db.example.com",
		"database":   "mydb",
		"port":       "3306",
		"note_value": "test_note",
		"password":   "test_password",
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
		"password":   "",
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
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "testuser"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "hostname", "db.example.com"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "database", "mydb"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "port", "3306"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.label", "test_section"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "section.0.field.0.label", "test_field"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "note_value", "test_note"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password", "test_password"),
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
					resource.TestCheckResourceAttr("onepassword_item.test_item", "password", ""),
				),
			},
		},
	})
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
