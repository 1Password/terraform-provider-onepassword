package provider

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )

// func TestAccExampleResource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Read testing
// 			{
// 				Config: testAccExampleResourceConfig("one"),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "one"),
// 					resource.TestCheckResourceAttr("scaffolding_example.test", "defaulted", "example value when not configured"),
// 					resource.TestCheckResourceAttr("scaffolding_example.test", "id", "example-id"),
// 				),
// 			},
// 			// ImportState testing
// 			{
// 				ResourceName:      "scaffolding_example.test",
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				// This is not normally necessary, but is here because this
// 				// example code does not have an actual upstream service.
// 				// Once the Read method is able to refresh information from
// 				// the upstream service, this can be removed.
// 				ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
// 			},
// 			// Update and Read testing
// 			{
// 				Config: testAccExampleResourceConfig("two"),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
// 				),
// 			},
// 			// Delete testing automatically occurs in TestCase
// 		},
// 	})
// }

// func testAccDataBaseResourceConfig(configurableAttribute string) string {
// 	return fmt.Sprintf(`
// 	resource "onepassword_item" "demo-database" {
// 		vault = data.onepassword_vault.acceptance-tests.uuid

// 		title = "Demo TF Resource - Database"
// 		category = "database"

// 		type = "MySQL"
// 		hostname = "http://my.fancydomain.com"
// 		port = "3000"
// 		database = "badass"
// 		username = "theboss"
// 		password_recipe {}
// 	}
// `, configurableAttribute)
// }
