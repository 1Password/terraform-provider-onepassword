package provider

import (
	"fmt"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccItemDataSourceSections(t *testing.T) {
	expectedItem := generateItemWithSections()
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
				Config: testAccProviderConfig(testServer.URL) + testAccItemDataSourceConfig(expectedItem.Vault.ID, expectedItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_item.test", "id", fmt.Sprintf("vaults/%s/items/%s", expectedVault.ID, expectedItem.ID)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "vault", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "uuid", expectedItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", string(expectedItem.Category)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "url", string(expectedItem.URLs[0].URL)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "tags.0", string(expectedItem.Tags[0])),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.id", expectedItem.Sections[0].ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.label", expectedItem.Sections[0].Label),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.field.0.label", expectedItem.Fields[0].Label),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.field.0.value", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.field.0.type", string(expectedItem.Fields[0].Type)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.field.0.purpose", string(expectedItem.Fields[0].Purpose)),
				),
			},
		},
	})
}

func TestAccItemDataSourceDatabase(t *testing.T) {
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
				Config: testAccProviderConfig(testServer.URL) + testAccItemDataSourceConfig(expectedItem.Vault.ID, expectedItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_item.test", "id", fmt.Sprintf("vaults/%s/items/%s", expectedVault.ID, expectedItem.ID)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "vault", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "uuid", expectedItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", string(expectedItem.Category)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "url", string(expectedItem.URLs[0].URL)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "tags.0", string(expectedItem.Tags[0])),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", expectedItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "hostname", expectedItem.Fields[2].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "database", expectedItem.Fields[3].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "port", expectedItem.Fields[4].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "type", expectedItem.Fields[5].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "note_value", expectedItem.Fields[6].Value),
				),
			},
		},
	})
}

func generateBaseItem() onepassword.Item {
	item := onepassword.Item{}
	item.ID = "rix6gwgpuyog4gqplegvrp3dbm"
	item.Vault.ID = "gs2jpwmahszwq25a7jiw45e4je"
	item.Category = "CUSTOM"
	item.Title = "test item"
	item.URLs = []onepassword.ItemURL{
		{
			Primary: true,
			URL:     "some_url.com",
		},
	}
	item.Tags = []string{
		"tag1",
	}

	return item
}

func generateItemWithSections() *onepassword.Item {
	item := generateBaseItem()
	section := &onepassword.ItemSection{
		ID:    "1234",
		Label: "Test Section",
	}
	item.Sections = append(item.Sections, section)
	item.Fields = append(item.Fields, &onepassword.ItemField{
		ID:      "23456",
		Type:    "STRING",
		Label:   "Secret Information",
		Value:   "Password123",
		Section: section,
	})

	return &item
}

func generateDatabaseItem() *onepassword.Item {
	item := generateBaseItem()
	item.Fields = generateDatabaseFields()

	return &item
}

func generateDatabaseFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test_user",
		},
		{
			Label: "password",
			Value: "test_password",
		},
		{
			Label: "hostname",
			Value: "test_host",
		},
		{
			Label: "database",
			Value: "test_database",
		},
		{
			Label: "port",
			Value: "test_port",
		},
		{
			Label: "type",
			Value: "test_type",
		},
		{
			Label:   "notes",
			Value:   "note_value",
			Purpose: "NOTES",
		},
	}
	return fields
}
func testAccItemDataSourceConfig(vault, uuid string) string {
	return fmt.Sprintf(`
data "onepassword_item" "test" {
  vault = "%s"
  uuid = "%s"
}`, vault, uuid)
}
