package integration

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
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
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/sections"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/ssh"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/uuid"
)

const testVaultID = "bbucuyq2nn4fozygwttxwizpcy"

type itemDataSourceTestCase struct {
	name                 string
	item                 testItem
	itemDataSourceConfig tfconfig.ItemDataSource
}

type testItem struct {
	Title string
	UUID  string
	Attrs map[string]string
}

var testItems = map[model.ItemCategory]testItem{
	model.Login: {
		Title: "Test Login",
		UUID:  "5axoqbjhbx3u7wqmersrg6qnqy",
		Attrs: map[string]string{
			"category": "login",
			"username": "testUsername",
			"password": "testPassword",
			"url":      "www.example.com",
		},
	},
	model.Password: {
		Title: "Test Password",
		UUID:  "axoqeauq7ilndgdpimb4j4dwhi",
		Attrs: map[string]string{
			"category": "password",
			"password": "testPassword",
		},
	},
	model.Database: {
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
	model.SecureNote: {
		Title: "Test Secure Note",
		UUID:  "5xbca3eblv5kxkszrbuhdame4a",
		Attrs: map[string]string{
			"category":   "secure_note",
			"note_value": "This is a test secure note for terraform-provider-onepassword",
		},
	},
	model.Document: {
		Title: "Test Document",
		UUID:  "p6uyugpmxo6zcxo5fdfctet7xa",
		Attrs: map[string]string{
			"category":              "document",
			"file.0.name":           "test.txt",
			"file.0.content":        "This is a test",
			"file.0.content_base64": "VGhpcyBpcyBhIHRlc3Q=",
		},
	},
	model.SSHKey: {
		Title: "Test SSH Key",
		UUID:  "5dbnxvhcknslz4mcaz7lobzt6i",
		Attrs: map[string]string{
			"category": "ssh_key",
		},
	},
	model.APICredential: {
		Title: "Test API Credential",
		UUID:  "fmx3tsd5v5bew74o3v2z2a7wf4",
		Attrs: map[string]string{
			"category":   "api_credential",
			"credential": "testCredential",
			"username":   "testAPICredential",
			"hostname":   "testHostname",
			"type":       "bearer",
			"valid_from": "2026-01-01",
			"filename":   "testFilename",
		},
	},
}

func TestAccItemDataSource(t *testing.T) {
	t.Parallel()

	createTestCase := func(name string, item testItem, identifierParam string, identifierValue string) itemDataSourceTestCase {
		return itemDataSourceTestCase{
			name: name,
			item: item,
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Params: map[string]string{
					identifierParam: identifierValue,
					"vault":         testVaultID,
				},
			},
		}
	}

	itemTypes := []struct {
		category model.ItemCategory
		name     string
	}{
		{model.Login, "Login"},
		{model.Password, "Password"},
		{model.Database, "Database"},
		{model.SecureNote, "SecureNote"},
		{model.Document, "Document"},
		{model.SSHKey, "SSHKey"},
		{model.APICredential, "APICredential"},
	}

	var testCases []itemDataSourceTestCase

	// Create test cases for each item type with both title and UUID lookup methods
	for _, itemType := range itemTypes {
		item := testItems[itemType.category]
		testCases = append(testCases,
			createTestCase(itemType.name+"ByTitle", item, "title", item.Title),
			createTestCase(itemType.name+"ByUUID", item, "uuid", item.UUID),
		)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dataSourceBuilder := tfconfig.CreateConfigBuilder()

			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "title", tc.item.Title),
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "uuid", tc.item.UUID),
			}

			for attr, expectedValue := range tc.item.Attrs {
				checks = append(checks, resource.TestCheckResourceAttr("data.onepassword_item.test_item", attr, expectedValue))
			}

			// Validate SSH keys
			if tc.item.Attrs["category"] == "ssh_key" {
				checks = append(checks, resource.TestCheckFunc(func(s *terraform.State) error {
					item, ok := s.RootModule().Resources["data.onepassword_item.test_item"]
					if !ok {
						return fmt.Errorf("resource not found in state")
					}

					publicKey := item.Primary.Attributes["public_key"]
					privateKey := item.Primary.Attributes["private_key"]

					return ssh.ValidateSSHKeys(publicKey, privateKey)
				}))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderConfig(),
						tfconfig.ItemDataSourceConfig(tc.itemDataSourceConfig.Params),
					),
					Check: resource.ComposeAggregateTestCheckFunc(checks...),
				}},
			})
		})
	}
}

func TestAccItemDataSource_NotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		item                 testItem
		itemDataSourceConfig tfconfig.ItemDataSource
	}{
		{
			name: "ByTitle",
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Params: map[string]string{
					"title": "invalid-title",
					"vault": testVaultID,
				},
			},
		},
		{
			name: "ByUUID",
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Params: map[string]string{
					"uuid":  "invalid-uuid",
					"vault": testVaultID,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dataSourceBuilder := tfconfig.CreateConfigBuilder()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderConfig(),
						tfconfig.ItemDataSourceConfig(tc.itemDataSourceConfig.Params),
					),
					ExpectError: regexp.MustCompile(`Unable to read item`),
				}},
			})
		})
	}
}

func TestAccItemDataSource_DetectManualChanges(t *testing.T) {
	t.Parallel()

	// Generate unique identifier for this test run to avoid conflicts in parallel execution
	uniqueID := uuid.New().String()
	var itemUUID string

	item := testItemsToCreate[model.Login]
	initialAttrs := maps.Clone(item.Attrs)
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
		"title":    initialAttrs["title"],
		"category": "login",
		"url":      []string{},
		"tags":     []string{},
		"section":  []map[string]any{},
	}

	// Initial data source read checks
	initialReadChecks := []resource.TestCheckFunc{
		logStep(t, "INITIAL_READ"),
		uuidutil.CaptureItemUUID(t, "data.onepassword_item.test_item", &itemUUID),
	}
	bcInitial := checks.BuildItemChecks("data.onepassword_item.test_item", initialAttrs)
	initialReadChecks = append(initialReadChecks, bcInitial...)

	// Build check function to manually update the item
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

	// Build checks for updated data source read
	updatedReadChecks := []resource.TestCheckFunc{
		logStep(t, "READ_AFTER_UPDATE"),
	}
	bcUpdated := checks.BuildItemChecks("data.onepassword_item.test_item", updatedAttrs)
	updatedReadChecks = append(updatedReadChecks, bcUpdated...)

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

	// Build checks for reading after field removal
	removedFieldsReadChecks := []resource.TestCheckFunc{
		logStep(t, "READ_AFTER_REMOVAL"),
	}
	bcRemoved := checks.BuildItemChecks("data.onepassword_item.test_item", removedAttrs)
	removedFieldsReadChecks = append(removedFieldsReadChecks, bcRemoved...)

	// Verify that username is either not present (SDK) or empty (Connect)
	removedFieldsReadChecks = append(removedFieldsReadChecks, resource.TestCheckFunc(func(s *terraform.State) error {
		item, ok := s.RootModule().Resources["data.onepassword_item.test_item"]
		if !ok {
			return fmt.Errorf("resource not found in state")
		}

		username, exists := item.Primary.Attributes["username"]
		if exists {
			// If username exists, it should be empty (Connect behavior)
			if username != "" {
				return fmt.Errorf("expected username to be empty or not present, got %q", username)
			}
		}
		// If username doesn't exist, that's also valid (SDK behavior)
		return nil
	}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create item using resource
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
				),
			},
			// Read item with data source
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": fmt.Sprintf("%v", initialAttrs["title"]),
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(initialReadChecks...),
			},
			// Manually update item
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": fmt.Sprintf("%v", initialAttrs["title"]),
					}),
				),
				Check:              updateItemOutsideTerraform(),
				ExpectNonEmptyPlan: true,
			},
			// Data source should read the updated values
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": fmt.Sprintf("%v", initialAttrs["title"]),
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(updatedReadChecks...),
			},
			// Manually remove fields
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": fmt.Sprintf("%v", initialAttrs["title"]),
					}),
				),
				Check:              removeFieldsOutsideTerraform(),
				ExpectNonEmptyPlan: true,
			},
			// Data source should read the removed fields
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, initialAttrs),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": fmt.Sprintf("%v", initialAttrs["title"]),
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(removedFieldsReadChecks...),
			},
		},
	})
}

func TestAccItemDataSourceSectionMap(t *testing.T) {
	t.Parallel()

	itemTitle := "Test Item with Sections"

	// Build checks for section_map
	sectionMapChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.onepassword_item.test_item", "section_map.%", "2"),
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Credentials.id"),
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Database Config.id"),
		// Field access via field_map
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Credentials.field_map.api_key.id"),
		resource.TestCheckResourceAttr("data.onepassword_item.test_item", "section_map.Credentials.field_map.api_key.type", "CONCEALED"),
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Credentials.field_map.api_key.value"),
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Credentials.field_map.api_secret.value"),
		// File in Database Config section
		resource.TestCheckResourceAttr("data.onepassword_item.test_item", "section_map.Database Config.file_map.test.txt.content", "This is a test"),
		resource.TestCheckResourceAttr("data.onepassword_item.test_item", "section_map.Database Config.file_map.test.txt.content_base64", "VGhpcyBpcyBhIHRlc3Q="),
		resource.TestCheckResourceAttrSet("data.onepassword_item.test_item", "section_map.Database Config.file_map.test.txt.id"),

		// Verify section_map matches section list
		resource.TestCheckFunc(func(s *terraform.State) error {
			item, ok := s.RootModule().Resources["data.onepassword_item.test_item"]
			if !ok {
				return fmt.Errorf("resource not found in state")
			}

			// Collect section IDs from the list
			var listSectionIDs []string
			for i := 0; ; i++ {
				id := item.Primary.Attributes[fmt.Sprintf("section.%d.id", i)]
				if id == "" {
					break
				}
				listSectionIDs = append(listSectionIDs, id)
			}

			// Collect section IDs from the map
			mapSectionIDs := []string{
				item.Primary.Attributes["section_map.Credentials.id"],
				item.Primary.Attributes["section_map.Database Config.id"],
			}

			// Compare slices (order-independent)
			slices.Sort(listSectionIDs)
			slices.Sort(mapSectionIDs)
			if !slices.Equal(listSectionIDs, mapSectionIDs) {
				return fmt.Errorf("section IDs mismatch: list=%v, map=%v", listSectionIDs, mapSectionIDs)
			}

			return nil
		}),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemDataSourceConfig(map[string]string{
						"vault": testVaultID,
						"title": itemTitle,
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(sectionMapChecks...),
			},
		},
	})
}

func TestAccItemDataSource_VaultName(t *testing.T) {
	t.Parallel()

	createTestCase := func(name string, item testItem, identifierParam string, identifierValue string) itemDataSourceTestCase {
		return itemDataSourceTestCase{
			name: name,
			item: item,
			itemDataSourceConfig: tfconfig.ItemDataSource{
				Params: map[string]string{
					identifierParam: identifierValue,
					"vault":         "terraform-provider-acceptance-tests",
				},
			},
		}
	}

	itemTypes := []struct {
		category model.ItemCategory
		name     string
	}{
		{model.Login, "Login"},
	}

	var testCases []itemDataSourceTestCase

	// Create test cases for each item type with both title and UUID lookup methods
	for _, itemType := range itemTypes {
		item := testItems[itemType.category]
		testCases = append(testCases,
			createTestCase(itemType.name+"ByTitle", item, "title", item.Title),
			createTestCase(itemType.name+"ByUUID", item, "uuid", item.UUID),
		)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dataSourceBuilder := tfconfig.CreateConfigBuilder()

			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "title", tc.item.Title),
				resource.TestCheckResourceAttr("data.onepassword_item.test_item", "uuid", tc.item.UUID),
			}

			for attr, expectedValue := range tc.item.Attrs {
				checks = append(checks, resource.TestCheckResourceAttr("data.onepassword_item.test_item", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderConfig(),
						tfconfig.ItemDataSourceConfig(tc.itemDataSourceConfig.Params),
					),
					Check: resource.ComposeAggregateTestCheckFunc(checks...),
				}},
			})
		})
	}
}
