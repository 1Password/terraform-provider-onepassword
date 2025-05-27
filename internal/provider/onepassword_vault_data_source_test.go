package provider

import (
	"fmt"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
	"github.com/stretchr/testify/require"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVaultDataSource(t *testing.T) {
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
				Config: testAccProviderConfig(testServer.URL) + testAccVaultDataSourceConfig(expectedItem.Vault.ID),
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

func TestOnePasswordVaultDataSourceModel_FromVault(t *testing.T) {
	vault := &model.Vault{
		ID:          "test-id",
		Title:       "test-name",
		Description: "test-description",
	}

	ds := &OnePasswordVaultDataSourceModel{}
	ds.FromVault(vault)

	require.Equal(t, fmt.Sprintf("vaults/%s", vault.ID), ds.ID)
	require.Equal(t, vault.ID, ds.UUID)
	require.Equal(t, vault.Title, ds.Name)
	require.Equal(t, vault.Description, ds.Description)
}
