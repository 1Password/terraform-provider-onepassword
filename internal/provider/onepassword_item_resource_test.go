package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func TestAccItemResourceDatabase(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
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
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
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
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
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
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
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
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
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

func TestAccItemResourceDocument(t *testing.T) {
	expectedItem := generateDocumentItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "VaultName",
		Description: "This vault will be retrieved for testing",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccDocumentResourceConfig(expectedItem),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

func TestAccItemResource_PasswordWriteOnly(t *testing.T) {
	expectedItem := generatePasswordItem()
	expectedVault := model.Vault{
		ID:   expectedItem.VaultID,
		Name: "VaultName",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test read
				Config: testAccProviderConfig(testServer.URL) + testAccPasswordWriteOnlyResourceConfig(expectedItem, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "password"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "password_wo"),
				),
			},
		},
	})
}

func TestAccItemResource_PasswordWriteOnlyAttributes(t *testing.T) {
	expectedItem := generatePasswordItem()
	expectedVault := model.Vault{
		ID:   expectedItem.VaultID,
		Name: "VaultName",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccPasswordWriteOnlyMissingVersionConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"password_wo_version\" must be specified when \"password_wo\" is"),
			},
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccPasswordWriteOnlyMissingPasswordConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"password_wo\" must be specified when \"password_wo_version\" is"),
			},
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccPasswordWriteOnlyConflictPasswordConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"password\" cannot be specified when \"password_wo\" is specified"),
			},
		},
	})
}

func TestAccItemResource_NoteValueWriteOnly(t *testing.T) {
	expectedItem := generateSecureNoteItem()
	expectedVault := model.Vault{
		ID:   expectedItem.VaultID,
		Name: "VaultName",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccNoteValueWriteOnlyResourceConfig(expectedItem, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "note_value_wo_version", "1"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "note_value"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "note_value_wo"),
				),
			},
		},
	})
}

func TestAccItemResource_NoteValueWriteOnlyAttributes(t *testing.T) {
	expectedItem := generateSecureNoteItem()
	expectedVault := model.Vault{
		ID:   expectedItem.VaultID,
		Name: "VaultName",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccNoteValueWriteOnlyMissingVersionConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"note_value_wo_version\" must be specified when \"note_value_wo\" is"),
			},
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccNoteValueWriteOnlyMissingNoteValueConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"note_value_wo\" must be specified when \"note_value_wo_version\" is"),
			},
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccNoteValueWriteOnlyConflictNoteValueConfig(expectedItem),
				ExpectError: regexp.MustCompile("Attribute \"note_value\" cannot be specified when \"note_value_wo\" is specified"),
			},
		},
	})
}

func testAccDataBaseResourceConfig(expectedItem *model.Item) string {
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
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[2].Value, expectedItem.Fields[3].Value, expectedItem.Fields[4].Value, expectedItem.Fields[5].Value)
}

func testAccNoteValueWriteOnlyResourceConfig(expectedItem *model.Item, version string) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  note_value_wo = <<EOT
%s
EOT
  note_value_wo_version = "%s"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), strings.TrimSuffix(expectedItem.Fields[0].Value, "\n"), version)
}

func testAccNoteValueWriteOnlyMissingVersionConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  note_value_wo = <<EOT
%s
EOT
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), strings.TrimSuffix(expectedItem.Fields[0].Value, "\n"))
}

func testAccNoteValueWriteOnlyMissingNoteValueConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  note_value_wo_version = "1"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)))
}

func testAccNoteValueWriteOnlyConflictNoteValueConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  note_value = <<EOT
%s
EOT
  note_value_wo = <<EOT
%s
EOT
  note_value_wo_version = "1"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[0].Value)
}

func testAccPasswordResourceConfig(expectedItem *model.Item) string {
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
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value)
}

func testAccPasswordWriteOnlyResourceConfig(expectedItem *model.Item, version string) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password_wo = "%s"
  password_wo_version = "%s"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[1].Value, version)
}

func testAccPasswordWriteOnlyMissingVersionConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password_wo = "%s"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[1].Value)
}

func testAccPasswordWriteOnlyMissingPasswordConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password_wo_version = "1"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value)
}

func testAccPasswordWriteOnlyConflictPasswordConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  username = "%s"
  password = "%s"
  password_wo = "%s"
  password_wo_version = "1"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.Fields[1].Value, expectedItem.Fields[1].Value)
}

func testAccLoginResourceConfig(expectedItem *model.Item) string {
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
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Fields[0].Value, expectedItem.URLs[0].URL)
}

func testAccSecureNoteResourceConfig(expectedItem *model.Item) string {
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
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), strings.TrimSuffix(expectedItem.Fields[0].Value, "\n"))
}

func testAccDocumentResourceConfig(expectedItem *model.Item) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test-document" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)))
}

func testAccResourceWithSectionsConfig(expectedItem *model.Item) string {
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
		expectedItem.VaultID,
		expectedItem.Title,
		strings.ToLower(string(expectedItem.Category)),
		expectedItem.Sections[0].Label,
		expectedItem.Fields[0].Label,
		expectedItem.Fields[0].Value,
	)
}
