package provider

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccItemDataSourceSections(t *testing.T) {
	expectedItem := generateItemWithSections()
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
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
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", expectedItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "hostname", expectedItem.Fields[2].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "database", expectedItem.Fields[3].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "port", expectedItem.Fields[4].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "type", expectedItem.Fields[5].Value),
				),
			},
		},
	})
}

func TestAccItemLoginDatabase(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", expectedItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "url", expectedItem.URLs[0].URL),
				),
			},
		},
	})
}

func TestAccItemPasswordDatabase(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", expectedItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemDocument(t *testing.T) {
	expectedItem := generateDocumentItem()
	expectedVault := op.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	first_content, err := expectedItem.Files[0].Content()
	if err != nil {
		t.Fatalf("Error getting content of first file: %v", err)
	}

	second_content, err := expectedItem.Files[1].Content()
	if err != nil {
		t.Fatalf("Error getting content of second file: %v", err)
	}

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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.0.id", expectedItem.Files[0].ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.0.name", expectedItem.Files[0].Name),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.0.content", string(first_content)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.1.id", expectedItem.Files[1].ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.1.name", expectedItem.Files[1].Name),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "file.1.content_base64", base64.StdEncoding.EncodeToString(second_content)),
				),
			},
		},
	})
}

func TestAccItemLoginWithFiles(t *testing.T) {
	expectedItem := generateLoginItemWithFiles()
	expectedVault := op.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	first_content, err := expectedItem.Files[0].Content()
	if err != nil {
		t.Fatalf("Error getting content of first file: %v", err)
	}

	second_content, err := expectedItem.Files[1].Content()
	if err != nil {
		t.Fatalf("Error getting content of second file: %v", err)
	}

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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.0.id", expectedItem.Files[0].ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.0.name", expectedItem.Files[0].Name),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.0.content", string(first_content)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.1.id", expectedItem.Files[1].ID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.1.name", expectedItem.Files[1].Name),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "section.0.file.1.content_base64", base64.StdEncoding.EncodeToString(second_content)),
				),
			},
		},
	})
}

func TestAccItemSSHKey(t *testing.T) {
	expectedItem := generateSSHKeyItem()
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "private_key", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "public_key", expectedItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemDataSourceApiCredential(t *testing.T) {
	expectedItem := generateApiCredentialItem()
	expectedVault := op.Vault{
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
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", strings.ToLower(string(expectedItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", expectedItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "credential", expectedItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "type", expectedItem.Fields[2].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "filename", expectedItem.Fields[3].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "valid_from", expectedItem.Fields[4].Value),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "hostname", expectedItem.Fields[5].Value),
				),
			},
		},
	})
}

func testAccItemDataSourceConfig(vault, uuid string) string {
	return fmt.Sprintf(`
data "onepassword_item" "test" {
  vault = "%s"
  uuid = "%s"
}`, vault, uuid)
}
