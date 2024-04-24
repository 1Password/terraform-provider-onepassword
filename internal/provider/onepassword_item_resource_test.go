package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccItemResourceDatabase(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := onepassword.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "VaultName",
		Description: "This vault will be retrieved for testing",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccDataBaseResourceConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					// verify local values
					resource.TestCheckResourceAttr("onepassword_item.test-database", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "hostname", expectedItem.Fields[2].Value),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "database", expectedItem.Fields[3].Value),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "port", expectedItem.Fields[4].Value),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "type", expectedItem.Fields[5].Value),
					resource.TestCheckResourceAttrSet("onepassword_item.test-database", "password"),
				),
			},
		},
	})
}

func TestAccItemResourcePassword(t *testing.T) {
	expectedItem := generatePasswordItem()
	expectedVault := onepassword.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "VaultName",
		Description: "This vault will be retrieved for testing",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccPasswordResourceConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					// verify local values
					resource.TestCheckResourceAttr("onepassword_item.test-database", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttrSet("onepassword_item.test-database", "password"),
				),
			},
		},
	})
}

func testAccDataBaseResourceConfig(expectedItem *onepassword.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}	
resource "onepassword_item" "test-database" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password_recipe {}
  hostname = "%s"
  database = "%s"
  port = "%s"
  type = "%s"
}`, expectedItem.Vault.ID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[2].Value, expectedItem.Fields[3].Value, expectedItem.Fields[4].Value, expectedItem.Fields[5].Value)
}

func testAccPasswordResourceConfig(expectedItem *onepassword.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}	
resource "onepassword_item" "test-database" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password_recipe {}
}`, expectedItem.Vault.ID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value)
}

// func testAccLoginResourceConfig(expectedItem *onepassword.Item) string {
// 	return fmt.Sprintf(`

// data "onepassword_vault" "acceptance-tests" {
// 	uuid = "%s"
// }
// resource "onepassword_item" "test-database" {
//   vault = data.onepassword_vault.acceptance-tests.uuid
//   title = "%s"
//   category = "%s"
//   username = "%s"
//   password_recipe {}
// }`, expectedItem.Vault.ID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value)
// }
