package utils

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestConfig holds configuration for e2e tests
type TestConfig struct {
	ServiceAccountToken string
}

// GetTestConfig retrieves test configuration from environment variables
func GetTestConfig() (*TestConfig, error) {
	config := &TestConfig{
		ServiceAccountToken: os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"),
	}

	if config.ServiceAccountToken == "" {
		return nil, fmt.Errorf("OP_SERVICE_ACCOUNT_TOKEN environment variable is required")
	}

	return config, nil
}

// GetProviderConfig returns the provider configuration for tests
func GetProviderConfig(config *TestConfig) string {
	return fmt.Sprintf(`
provider "onepassword" {
  service_account_token = "%s"
}`, config.ServiceAccountToken)
}

// ValidateResourceExists checks if a resource exists in the state
func ValidateResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has no ID", resourceName)
		}

		return nil
	}
}

// ValidateResourceAttribute checks if a resource attribute has the expected value
func ValidateResourceAttribute(resourceName, attribute, expectedValue string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttr(resourceName, attribute, expectedValue)
}

// ValidateResourceAttributeSet checks if a resource attribute is set
func ValidateResourceAttributeSet(resourceName, attribute string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrSet(resourceName, attribute)
}

// TestAccItemDataSourceConfig generates a Terraform configuration for testing the onepassword_item data source
func TestAccItemDataSourceConfig(config *TestConfig, vaultID, identifierType, identifierValue string) string {
	return fmt.Sprintf(`
%s
data "onepassword_item" "test" {
  %s    = "%s"
  vault = "%s"
}
`, GetProviderConfig(config), identifierType, identifierValue, vaultID)
}

