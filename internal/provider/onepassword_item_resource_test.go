package provider

import (
	"fmt"
	"strings"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccItemResourceDatabase(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := op.Vault{
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
	expectedVault := op.Vault{
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

func TestAccItemResourceLogin(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := op.Vault{
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
				Config: testAccProviderConfig(testServer.URL) + testAccLoginResourceConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					// verify local values
					resource.TestCheckResourceAttr("onepassword_item.test-database", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "url", expectedItem.URLs[0].URL),
					resource.TestCheckResourceAttrSet("onepassword_item.test-database", "password"),
				),
			},
		},
	})
}

func TestAccItemResourceSecureNote(t *testing.T) {
	expectedItem := generateSecureNoteItem()
	expectedVault := op.Vault{
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
				Config: testAccProviderConfig(testServer.URL) + testAccSecureNoteResourceConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					// verify local values
					resource.TestCheckResourceAttr("onepassword_item.test-secure-note", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test-secure-note", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test-secure-note", "note_value", expectedItem.Fields[0].Value),
					resource.TestCheckNoResourceAttr("onepassword_item.test-secure-note", "password"),
				),
			},
		},
	})
}

func TestAccItemResourceWithSections(t *testing.T) {
	expectedItem := generateItemWithSections()
	expectedVault := op.Vault{
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
				Config: testAccProviderConfig(testServer.URL) + testAccResourceWithSectionsConfig(expectedItem),
				Check: resource.ComposeAggregateTestCheckFunc(
					// verify local values
					resource.TestCheckResourceAttr("onepassword_item.test-database", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttrSet("onepassword_item.test-database", "password"),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "section.0.label", expectedItem.Sections[0].Label),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "section.0.field.0.label", expectedItem.Fields[0].Label),
					resource.TestCheckResourceAttr("onepassword_item.test-database", "section.0.field.0.value", expectedItem.Fields[0].Value),
				),
			},
		},
	})
}

func testAccDataBaseResourceConfig(expectedItem *op.Item) string {
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

func testAccPasswordResourceConfig(expectedItem *op.Item) string {
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

func testAccLoginResourceConfig(expectedItem *op.Item) string {
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
  url = "%s"
}`, expectedItem.Vault.ID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.URLs[0].URL)
}

func testAccSecureNoteResourceConfig(expectedItem *op.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test-secure-note" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  note_value = <<EOT
%s
EOT
}`, expectedItem.Vault.ID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), strings.TrimSuffix(expectedItem.Fields[0].Value, "\n"))
}

func testAccResourceWithSectionsConfig(expectedItem *op.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test-database" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  password_recipe {}
  section {
	label = "%s"
	field {
	  label = "%s"
	  value = "%s"
	}
  }
}`,
		expectedItem.Vault.ID,
		expectedItem.Title,
		strings.ToLower(string(expectedItem.Category)),
		expectedItem.Sections[0].Label,
		expectedItem.Fields[0].Label,
		expectedItem.Fields[0].Value,
	)
}
