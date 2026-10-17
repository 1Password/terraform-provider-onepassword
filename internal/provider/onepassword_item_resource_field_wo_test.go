package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

func TestAccItemResource_SectionFieldWriteOnly(t *testing.T) {
	expectedItem := generateItemWithSections()
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
				Config: testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyConfig(expectedItem, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "section.0.label", expectedItem.Sections[0].Label),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "section.0.field.0.label", expectedItem.Fields[0].Label),
					resource.TestCheckResourceAttr("onepassword_item.test_wo", "section.0.field.0.value_wo_version", "1"),
					resource.TestCheckResourceAttrSet("onepassword_item.test_wo", "section.0.field.0.id"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "section.0.field.0.value"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", "section.0.field.0.value_wo"),
				),
			},
		},
	})
}

func TestAccItemResource_SectionMapFieldWriteOnly(t *testing.T) {
	expectedItem := generateItemWithSections()
	expectedVault := model.Vault{
		ID:   expectedItem.VaultID,
		Name: "VaultName",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	section := expectedItem.Sections[0].Label
	field := expectedItem.Fields[0].Label
	prefix := fmt.Sprintf("section_map.%s.field_map.%s", section, field)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccSectionMapFieldWriteOnlyConfig(expectedItem, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_wo", prefix+".value_wo_version", "1"),
					resource.TestCheckResourceAttrSet("onepassword_item.test_wo", prefix+".id"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", prefix+".value"),
					resource.TestCheckNoResourceAttr("onepassword_item.test_wo", prefix+".value_wo"),
				),
			},
		},
	})
}

func TestAccItemResource_SectionFieldWriteOnlyAttributes(t *testing.T) {
	expectedItem := generateItemWithSections()
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
				Config:      testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyAttributesConfig(expectedItem, `value_wo = "secret"`),
				ExpectError: regexp.MustCompile(`Attribute "section\[0\].field\[0\].value_wo_version" must be specified when`),
			},
			{
				Config:      testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyAttributesConfig(expectedItem, `value_wo_version = 1`),
				ExpectError: regexp.MustCompile(`Attribute "section\[0\].field\[0\].value_wo" must be specified when`),
			},
			{
				Config: testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyAttributesConfig(expectedItem, `
      value            = "secret"
      value_wo         = "secret"
      value_wo_version = 1`),
				ExpectError: regexp.MustCompile(`Attribute "section\[0\].field\[0\].value" cannot be specified when`),
			},
			{
				Config: testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyAttributesConfig(expectedItem, `
      value_wo         = "secret"
      value_wo_version = 1
      password_recipe {}`),
				ExpectError: regexp.MustCompile(`Attribute "section\[0\].field\[0\].password_recipe" cannot be specified when`),
			},
			{
				Config: testAccProviderConfig(testServer.URL) + testAccSectionFieldWriteOnlyAttributesConfig(expectedItem, `
      type             = "MONTH_YEAR"
      value_wo         = "2024-01"
      value_wo_version = 1`),
				ExpectError: regexp.MustCompile("Invalid MONTH_YEAR format"),
			},
		},
	})
}

func testAccSectionFieldWriteOnlyConfig(expectedItem *model.Item, version string) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  password_recipe {}
  section {
    label = "%s"
    field {
      label            = "%s"
      value_wo         = "%s"
      value_wo_version = %s
    }
  }
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Sections[0].Label, expectedItem.Fields[0].Label, expectedItem.Fields[0].Value, version)
}

func testAccSectionMapFieldWriteOnlyConfig(expectedItem *model.Item, version string) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  password_recipe {}
  section_map = {
    "%s" = {
      field_map = {
        "%s" = {
          value_wo         = "%s"
          value_wo_version = %s
        }
      }
    }
  }
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Sections[0].Label, expectedItem.Fields[0].Label, expectedItem.Fields[0].Value, version)
}

// testAccSectionFieldWriteOnlyAttributesConfig renders one section field with the given
// attributes, for checking the validators on value_wo and value_wo_version.
func testAccSectionFieldWriteOnlyAttributesConfig(expectedItem *model.Item, fieldAttributes string) string {
	return fmt.Sprintf(`

data "onepassword_vault" "acceptance-tests" {
	uuid = "%s"
}
resource "onepassword_item" "test_wo" {
  vault = data.onepassword_vault.acceptance-tests.uuid
  title = "%s"
  category = "%s"
  section {
    label = "%s"
    field {
      label = "%s"
      %s
    }
  }
}`, expectedItem.VaultID, expectedItem.Title, strings.ToLower(string(expectedItem.Category)), expectedItem.Sections[0].Label, expectedItem.Fields[0].Label, fieldAttributes)
}
