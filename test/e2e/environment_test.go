package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
)

// environmentDataSourceConfig returns Terraform config for the onepassword_environment datasource. Private to this test.
func environmentDataSourceConfig(environmentID string) string {
	return fmt.Sprintf(`data "onepassword_environment" "test_environment" {
  environment_id = %q
}
`, environmentID)
}

func TestAccEnvironmentDataSource(t *testing.T) {
	t.Parallel()

	environmentID := os.Getenv("OP_TEST_ENVIRONMENT_ID")
	if environmentID == "" {
		t.Skip("OP_TEST_ENVIRONMENT_ID must be set for this test (1Password Environment ID from Developer > View Environments)")
	}

	dataSourceBuilder := tfconfig.CreateConfigBuilder()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: dataSourceBuilder(
				tfconfig.ProviderConfig(),
				func() string { return environmentDataSourceConfig(environmentID) },
			),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.onepassword_environment.test_environment", "id", "environments/"+environmentID),
				resource.TestCheckResourceAttr("data.onepassword_environment.test_environment", "environment_id", environmentID),
				resource.TestCheckResourceAttrSet("data.onepassword_environment.test_environment", "variables.%"),
				resource.TestCheckResourceAttrSet("data.onepassword_environment.test_environment", "metadata.#"),
				// metadata (list of objects) and variables (map) expose the same env vars in two shapes; their counts must match.
				resource.TestCheckResourceAttrPair(
					"data.onepassword_environment.test_environment", "metadata.#",
					"data.onepassword_environment.test_environment", "variables.%",
				),
			),
		}},
	})
}
