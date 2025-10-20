package integration

import (
	"fmt"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v2/e2e/utils"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// e2e testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"onepassword": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccVaultDataSourceByName(t *testing.T) {
	config, err := utils.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	expectedVault := op.Vault{
		ID:          "oogirevsqtgi4foiv66ntgowwm",
		Name:        "operator-acceptance-tests",
		Description: "This vault contains items to be used in Kubernetes Operator e2e tests.",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultDataSourceConfig(config, expectedVault.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "id", fmt.Sprintf("vaults/%s", expectedVault.ID)),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "uuid", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "name", expectedVault.Name),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "description", expectedVault.Description),
				),
			},
		},
	})
}

func TestAccVaultDataSourceByUUID(t *testing.T) {
	config, err := utils.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	expectedVault := op.Vault{
		ID:          "oogirevsqtgi4foiv66ntgowwm",
		Name:        "operator-acceptance-tests",
		Description: "This vault contains items to be used in Kubernetes Operator e2e tests.",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultDataSourceByUUIDConfig(config, expectedVault.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					utils.ValidateResourceExists("data.onepassword_vault.test"),
					utils.ValidateResourceAttributeSet("data.onepassword_vault.test", "name"),
					utils.ValidateResourceAttributeSet("data.onepassword_vault.test", "uuid"),
					utils.ValidateResourceAttributeSet("data.onepassword_vault.test", "description"),
				),
			},
		},
	})
}

func testAccVaultDataSourceConfig(config *utils.TestConfig, vaultName string) string {
	return fmt.Sprintf(`
%s

# Test reading a pre-existing vault by name
data "onepassword_vault" "test" {
 name = "%s"
}
`, utils.GetProviderConfig(config), vaultName)
}

func testAccVaultDataSourceByUUIDConfig(config *utils.TestConfig, vaultUUID string) string {
	return fmt.Sprintf(`
%s

# Test reading a pre-existing vault by UUID
data "onepassword_vault" "test" {
 uuid = "%s"
}
`, utils.GetProviderConfig(config), vaultUUID)
}
