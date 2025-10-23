// This package is used to provide helpers to generate terraform configuration files
package terraform

import (
	"fmt"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
)
type DataSourceConfigParams struct {
    TestConfig      *config.TestConfig
    DataSource  	string
    Vault           string
    IdentifierType  string
    IdentifierValue string
}

// Provider returns the provider configuration for tests
func Provider(config *config.TestConfig) string {
	return fmt.Sprintf(`
provider "onepassword" {
  service_account_token = "%s"
}`, config.ServiceAccountToken)
}

// DataSource returns the terraform configuration for a data source
func DataSource(config DataSourceConfigParams) string {
	var vaultLine string
	identifierLine := fmt.Sprintf(`%s = "%s"`, config.IdentifierType, config.IdentifierValue)

	if config.DataSource == "onepassword_item" {
		vaultLine = fmt.Sprintf(`vault = "%s"`, config.Vault)
	}

	return fmt.Sprintf(`
%s
data "%s" "test" {
%s
%s
}
`, Provider(config.TestConfig), config.DataSource, identifierLine, vaultLine)
}

