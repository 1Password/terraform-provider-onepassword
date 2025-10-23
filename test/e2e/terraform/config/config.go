// This package is used to provide helpers to generate terraform configuration files
package terraform

import (
	"fmt"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
)

// Provider returns the provider configuration for tests
func Provider(config *config.TestConfig) string {
	return fmt.Sprintf(`
provider "onepassword" {
  service_account_token = "%s"
}`, config.ServiceAccountToken)
}

func VaultDataSourceByName(config *config.TestConfig, vaultName string) string {
	return fmt.Sprintf(`
%s

# Test reading a pre-existing vault by name
data "onepassword_vault" "test" {
 name = "%s"
}
`, Provider(config), vaultName)
}

func VaultDataSourceByUUID(config *config.TestConfig, vaultUUID string) string {
	return fmt.Sprintf(`
%s

# Test reading a pre-existing vault by UUID
data "onepassword_vault" "test" {
 uuid = "%s"
}
`, Provider(config), vaultUUID)
}

func ItemDataSource(config *config.TestConfig, vaultID, identifierType, identifierValue string) string {
    return fmt.Sprintf(`
%s
data "onepassword_item" "test" {
  %s    = "%s"
  vault = "%s"
}
`, Provider(config), identifierType, identifierValue, vaultID)
}
