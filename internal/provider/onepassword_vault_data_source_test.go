package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func TestAccVaultDataSource(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccVaultDataSourceConfig(expectedItem.VaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "id", fmt.Sprintf("vaults/%s", expectedVault.ID)),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "uuid", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "description", expectedVault.Description),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "name", expectedVault.Name),
				),
			},
		},
	})
}

func testAccVaultDataSourceConfig(vault string) string {
	return fmt.Sprintf(`
data "onepassword_vault" "test" {
  uuid = "%s"
}`, vault)
}
