package integration

import (
	"fmt"
	"testing"

	op "github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/provider"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
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
	config, err := config.GetTestConfig()
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
				Config: tfconfig.VaultDataSourceByName(config, expectedVault.Name),
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
	config, err := config.GetTestConfig()
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
				Config: tfconfig.VaultDataSourceByUUID(config, expectedVault.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "name", expectedVault.Name),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "uuid", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_vault.test", "description", expectedVault.Description),
				),
			},
		},
	})
}


