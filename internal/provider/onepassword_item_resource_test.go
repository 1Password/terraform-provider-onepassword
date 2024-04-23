package provider

import (
	"fmt"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := onepassword.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccDataBaseResourceConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "one"),
					resource.TestCheckResourceAttr("scaffolding_example.test", "defaulted", "example value when not configured"),
					resource.TestCheckResourceAttr("scaffolding_example.test", "id", "example-id"),
				),
			},
			// // ImportState testing
			// {
			// 	ResourceName:      "scaffolding_example.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// This is not normally necessary, but is here because this
			// 	// example code does not have an actual upstream service.
			// 	// Once the Read method is able to refresh information from
			// 	// the upstream service, this can be removed.
			// 	ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			// },
			// // Update and Read testing
			// {
			// 	Config: testAccDataBaseResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
			// 	),
			// },
			// // Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDataBaseResourceConfig(expectedItem *onepassword.Item) string {
	return fmt.Sprintf(`
	data "onepassword_vault" "test-vault {
		uuid = "%s"
	}

	resource "onepassword_item" "test-database" {
		vault = data.onepassword_vault.test-vault.uuid

		title = %s
		category = %s

		username = %s
		password_recipe {}
		hostname = %s
		database = %s
		port = %s
		type = %s
	}
`, expectedItem.Vault.ID, expectedItem.Title, expectedItem.Category, expectedItem.Fields[0].Value, expectedItem.Fields[1].Value, expectedItem.Fields[2].Value, expectedItem.Fields[3].Value, expectedItem.Fields[4].Value)
}
