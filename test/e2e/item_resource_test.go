package integration

import (
	"fmt"
	"regexp"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type testItemResource struct {
	Title string
	Attrs map[string]string
}

var testItemsToCreate = map[op.ItemCategory]testItemResource{
	op.Login: {
		Title: "Test Login Create",
		Attrs: map[string]string{
			"vault":      "t7dnwbjh6nlyw475wl3m442sdi",
			"category":   "login",
			"username":   "testuser@example.com",
			"password":   "testPassword",
			"url":        "https://example.com",
			"note_value": "Test login note",
		},
	},
	op.Password: {
		Title: "Test Password Create",
		Attrs: map[string]string{
			"vault":    "bbucuyq2nn4fozygwttxwizpcy",
			"category": "password",
			"password": "testPassword",
		},
	},
	op.Database: {
		Title: "Test Database Create",
		Attrs: map[string]string{
			"vault":    "bbucuyq2nn4fozygwttxwizpcy",
			"category": "database",
			"username": "testUsername",
			"password": "testPassword",
			"database": "testDatabase",
			"port":     "3306",
			"type":     "mysql",
		},
	},
	op.SecureNote: {
		Title: "Test Secure Note Create",
		Attrs: map[string]string{
			"vault":      "bbucuyq2nn4fozygwttxwizpcy",
			"category":   "secure_note",
			"note_value": "This is a test secure note",
		},
	},
}

// func TestAccItemResourceCreate(t *testing.T) {
// 	testCases := []struct {
// 		category op.ItemCategory
// 		name     string
// 	}{
// 		{op.Login, "Login"},
// 		{op.Password, "Password"},
// 		{op.Database, "Database"},
// 		{op.SecureNote, "SecureNote"},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			item := testItemsToCreate[tc.category]
// 			resourceBuilder := tfconfig.CreateItemResourceConfigBuilder()

// 			config := make(map[string]string)
// 			config["title"] = item.Title
// 			for k, v := range item.Attrs {
// 				config[k] = v
// 			}

// 			checks := []resource.TestCheckFunc{
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
// 				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "id"),
// 				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "uuid"),
// 			}

// 			for attr, expectedValue := range item.Attrs {
// 				checks = append(checks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
// 			}

// 			resource.Test(t, resource.TestCase{
// 				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 				Steps: []resource.TestStep{
// 					{
// 						Config: resourceBuilder(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemResourceConfig(config),
// 						),
// 						Check: resource.ComposeAggregateTestCheckFunc(checks...),
// 					},
// 				},
// 			})
// 		})
// 	}
// }

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

// func TestAccItemResourceUpdate(t *testing.T) {
// 	testCases := []struct {
// 		category op.ItemCategory
// 		name     string
// 	}{
// 		{op.Login, "Login"},
// 		{op.Password, "Password"},
// 		{op.Database, "Database"},
// 		{op.SecureNote, "SecureNote"},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			item := testItemsToCreate[tc.category]

// 			initialConfig := make(map[string]string)
// 			initialConfig["title"] = item.Title
// 			for k, v := range item.Attrs {
// 				initialConfig[k] = v
// 			}

// 			updatedConfig := make(map[string]string)
// 			updatedConfig["title"] = item.Title
// 			for k, v := range item.Attrs {
// 				updatedConfig[k] = v
// 			}

// 			for k, v := range testItemsUpdatedAttrs[tc.category] {
// 				updatedConfig[k] = v
// 			}

// 			initialChecks := []resource.TestCheckFunc{
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
// 				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "id"),
// 			}
// 			for attr, expectedValue := range item.Attrs {
// 				initialChecks = append(initialChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
// 			}

// 			updatedChecks := []resource.TestCheckFunc{
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
// 			}
// 			for attr, expectedValue := range testItemsUpdatedAttrs[tc.category] {
// 				updatedChecks = append(updatedChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
// 			}

// 			resource.Test(t, resource.TestCase{
// 				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 				Steps: []resource.TestStep{
// 					{
// 						Config: tfconfig.CreateItemResourceConfigBuilder()(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemResourceConfig(initialConfig),
// 						),
// 						Check: resource.ComposeAggregateTestCheckFunc(initialChecks...),
// 					},
// 					{
// 						Config: tfconfig.CreateItemResourceConfigBuilder()(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemResourceConfig(updatedConfig),
// 						),
// 						Check: resource.ComposeAggregateTestCheckFunc(updatedChecks...),
// 					},
// 				},
// 			})
// 		})
// 	}
// }

// func TestAccItemResourceRead(t *testing.T) {
// 	testCases := []struct {
// 		category op.ItemCategory
// 		name     string
// 	}{
// 		{op.Login, "Login"},
// 		{op.Password, "Password"},
// 		{op.Database, "Database"},
// 		{op.SecureNote, "SecureNote"},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			item := testItems[tc.category]
// 			resourceBuilder := tfconfig.CreateItemResourceConfigBuilder()

// 			config := make(map[string]string)
// 			config["vault"] = testVaultID
// 			config["title"] = item.Title
// 			for k, v := range item.Attrs {
// 				config[k] = v
// 			}

// 			checks := []resource.TestCheckFunc{
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "uuid", item.UUID),
// 				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "id"),
// 			}
// 			for attr, expectedValue := range item.Attrs {
// 				checks = append(checks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
// 			}

// 			resource.Test(t, resource.TestCase{
// 				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 				Steps: []resource.TestStep{
// 					{
// 						Config: resourceBuilder(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemResourceConfig(config),
// 						),
// 						ResourceName:      "onepassword_item.test_item",
// 						ImportState:       true,
// 						ImportStateId:     fmt.Sprintf("vaults/%s/items/%s", testVaultID, item.UUID),
// 						ImportStateVerify: false,
// 						Check:             resource.ComposeAggregateTestCheckFunc(checks...),
// 					},
// 				},
// 			})
// 		})
// 	}
// }

// func TestAccItemResourceDelete(t *testing.T) {
// 	testCases := []struct {
// 		category op.ItemCategory
// 		name     string
// 	}{
// 		{op.Login, "Login"},
// 		{op.Password, "Password"},
// 		{op.Database, "Database"},
// 		{op.SecureNote, "SecureNote"},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			item := testItemsToCreate[tc.category]

// 			config := make(map[string]string)
// 			config["title"] = item.Title
// 			for k, v := range item.Attrs {
// 				config[k] = v
// 			}

// 			checks := []resource.TestCheckFunc{
// 				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
// 				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "id"),
// 			}

// 			resource.Test(t, resource.TestCase{
// 				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 				Steps: []resource.TestStep{
// 					{
// 						Config: tfconfig.CreateItemResourceConfigBuilder()(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemResourceConfig(config),
// 						),
// 						Check: resource.ComposeAggregateTestCheckFunc(checks...),
// 					},
// 					{
// 						Config: tfconfig.CreateItemResourceConfigBuilder()(
// 							tfconfig.ProviderConfig(),
// 						),
// 					},
// 					{
// 						Config: tfconfig.CreateItemResourceConfigBuilder()(
// 							tfconfig.ProviderConfig(),
// 							tfconfig.ItemDataSourceConfig(
// 								map[string]string{
// 									"vault": testVaultID,
// 									"title": item.Title,
// 								},
// 							),
// 						),
// 						ExpectError: regexp.MustCompile("Unable to read item"),
// 					},
// 				},
// 			})
// 		})
// 	}
// }

func TestAccItemResourceCRUD(t *testing.T) {
	testCases := []struct {
		category op.ItemCategory
		name     string
	}{
		{op.Login, "Login"},
		// {op.Password, "Password"},
		// {op.Database, "Database"},
		// {op.SecureNote, "SecureNote"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := testItemsToCreate[tc.category]

			// Create Config
			initialConfig := make(map[string]string)
			initialConfig["title"] = item.Title
			for k, v := range item.Attrs {
				initialConfig[k] = v
			}

			// Update Config
			updatedConfig := make(map[string]string)
			updatedConfig["title"] = item.Title
			for k, v := range item.Attrs {
				updatedConfig[k] = v
			}

			for k, v := range testItemsUpdatedAttrs[tc.category] {
				updatedConfig[k] = v
			}

			// Create Checks
			initialChecks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
				resource.TestCheckResourceAttrSet("onepassword_item.test_item", "id"),
			}
			for attr, expectedValue := range item.Attrs {
				initialChecks = append(initialChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
			}

			// Update Checks
			updatedChecks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "title", item.Title),
			}
			for attr, expectedValue := range testItemsUpdatedAttrs[tc.category] {
				updatedChecks = append(updatedChecks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create
					{
						Config: tfconfig.CreateItemResourceConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(initialConfig),
						),
						//Check: resource.ComposeAggregateTestCheckFunc(initialChecks...),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							resource.TestCheckFunc(func(s *terraform.State) error {
								t.Logf("CREATE")
								return nil
							}),
						}, initialChecks...)...),
					},
					// Read
					{
						Config: tfconfig.CreateItemResourceConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(initialConfig),
						),
						ResourceName:  "onepassword_item.test_item",
						ImportStateId: fmt.Sprintf("vaults/%s/items/%s", "t7dnwbjh6nlyw475wl3m442sdi", item.Title),
						Check: resource.ComposeAggregateTestCheckFunc(append([]resource.TestCheckFunc{
							resource.TestCheckFunc(func(s *terraform.State) error {
								t.Logf("READING")
								return nil
							}),
						}, initialChecks...)...),
					},
					// Update
					{
						Config: tfconfig.CreateItemResourceConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(updatedConfig),
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
						Config: tfconfig.CreateItemResourceConfigBuilder()(
							tfconfig.ProviderConfig(),
						),
						Check: resource.TestCheckFunc(func(s *terraform.State) error {
							t.Logf("DELETE")
							return nil
						}),
					},
					{
						Config: tfconfig.CreateItemResourceConfigBuilder()(
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
