package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/cleanup"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/uuid"
)

func TestAccVaultResource(t *testing.T) {
	if os.Getenv("OP_TEST_VAULT_MANAGEMENT") == "" {
		t.Skip("set OP_TEST_VAULT_MANAGEMENT to run destructive vault-management tests")
	}
	if os.Getenv("OP_CONNECT_HOST") != "" || os.Getenv("OP_CONNECT_TOKEN") != "" {
		t.Skip("vault management is not supported with 1Password Connect")
	}

	groupID := os.Getenv("OP_TEST_VAULT_GROUP_ID")
	name := "Terraform acceptance vault " + uuid.NewString()
	description := "Created by terraform-provider-onepassword acceptance tests"
	updatedDescription := "Updated by terraform-provider-onepassword acceptance tests"
	var vaultUUID string

	createConfig := tfconfig.CreateConfigBuilder()(
		tfconfig.ProviderConfig(),
		tfconfig.VaultResourceConfig(name, description, false, groupID, []string{"read_items"}),
	)
	updateConfig := tfconfig.CreateConfigBuilder()(
		tfconfig.ProviderConfig(),
		tfconfig.VaultResourceConfig(name, updatedDescription, false, groupID, []string{"read_items", "create_items"}),
	)
	if os.Getenv("OP_ACCOUNT") == "" {
		// Service accounts cannot update vault metadata, but can manage permissions
		// for vaults they created.
		updateConfig = tfconfig.CreateConfigBuilder()(
			tfconfig.ProviderConfig(),
			tfconfig.VaultResourceConfig(name, description, false, groupID, []string{"read_items", "create_items"}),
		)
	}

	steps := []resource.TestStep{{
		Config: createConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			uuidutil.CaptureItemUUID(t, "onepassword_vault.test_vault", &vaultUUID),
			cleanup.RegisterVault(t, &vaultUUID),
			resource.TestCheckResourceAttr("onepassword_vault.test_vault", "name", name),
			resource.TestCheckResourceAttr("onepassword_vault.test_vault", "description", description),
			resource.TestCheckResourceAttr("onepassword_vault.test_vault", "allow_admins_access", "false"),
		),
	}}

	if groupID != "" {
		steps = append(steps, resource.TestStep{
			Config: updateConfig,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("onepassword_vault.test_vault", "group_permissions.#", "1"),
				resource.TestCheckResourceAttr("onepassword_vault.test_vault", "description", func() string {
					if os.Getenv("OP_ACCOUNT") != "" {
						return updatedDescription
					}
					return description
				}()),
			),
		})
		descriptionAfterUpdate := description
		if os.Getenv("OP_ACCOUNT") != "" {
			descriptionAfterUpdate = updatedDescription
		}
		steps = append(steps, resource.TestStep{
			Config: tfconfig.CreateConfigBuilder()(
				tfconfig.ProviderConfig(),
				tfconfig.VaultResourceConfig(name, descriptionAfterUpdate, false, "", nil),
			),
			Check: resource.TestCheckResourceAttr("onepassword_vault.test_vault", "group_permissions.#", "0"),
		})
	} else if os.Getenv("OP_ACCOUNT") != "" {
		steps = append(steps, resource.TestStep{
			Config: updateConfig,
			Check:  resource.TestCheckResourceAttr("onepassword_vault.test_vault", "description", updatedDescription),
		})
	}

	steps = append(steps,
		resource.TestStep{
			ResourceName: "onepassword_vault.test_vault",
			ImportState:  true,
			ImportStateIdFunc: func(_ *terraform.State) (string, error) {
				if vaultUUID == "" {
					return "", fmt.Errorf("vault UUID was not captured")
				}
				return "vaults/" + vaultUUID, nil
			},
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"allow_admins_access", "group_permissions"},
		},
		resource.TestStep{
			Config: tfconfig.CreateConfigBuilder()(tfconfig.ProviderConfig()),
		},
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
