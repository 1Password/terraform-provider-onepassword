package integration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/provider"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// e2e testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"onepassword": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccVaultDataSource(t *testing.T) {
	serviceAccountToken, err := config.GetServiceAccountToken()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	expectedVaultAttrs := map[string]string{
		"description": "This vault contains the items that are used for 1Password Terraform Provider acceptance (e2e) tests.",
		"name":        "terraform-provider-acceptance-tests",
		"uuid":        "bbucuyq2nn4fozygwttxwizpcy",
	}

	testCases := []struct {
		name                  string
		identifierParam       string
		identifierValue       string
		expectedAttrs         map[string]string
		vaultDataSourceConfig tfconfig.VaultDataSource
	}{
		{
			name:            "ByName",
			identifierParam: "name",
			identifierValue: "terraform-provider-acceptance-tests",
			expectedAttrs:   expectedVaultAttrs,
			vaultDataSourceConfig: tfconfig.VaultDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"name": "terraform-provider-acceptance-tests",
				},
			},
		},
		{
			name:            "ByUUID",
			identifierParam: "uuid",
			identifierValue: "bbucuyq2nn4fozygwttxwizpcy",
			expectedAttrs:   expectedVaultAttrs,
			vaultDataSourceConfig: tfconfig.VaultDataSource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"uuid": "bbucuyq2nn4fozygwttxwizpcy",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataSourceBuilder := tfconfig.CreateItemDataSourceConfigBuilder()

			checks := make([]resource.TestCheckFunc, 0, len(tc.expectedAttrs))
			for attr, expectedValue := range tc.expectedAttrs {
				checks = append(checks, resource.TestCheckResourceAttr("data.onepassword_vault.test_vault", attr, expectedValue))
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: dataSourceBuilder(
						tfconfig.ProviderAuthWithServiceAccount(tc.vaultDataSourceConfig.Auth),
						tfconfig.VaultDataSourceConfig(tc.vaultDataSourceConfig.Params),
					),
					Check: resource.ComposeTestCheckFunc(checks...),
				}},
			})
		})
	}
}
