package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/e2e/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoginItemDataSourceByTitle(t *testing.T) {
	config, err := utils.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	vaultID := "t7dnwbjh6nlyw475wl3m442sdi"
	itemTitle := "Test Login"

	itemUUID := "dsrwv5dyacw4f7pdrfnmh36pne"
	username := "testUsername"
	password := "testPassword"


	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoginItemDataSourceByTitleConfig(config, itemTitle, vaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_item.test", "title", itemTitle),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", "login"),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "uuid", itemUUID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", username),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", password),
				),
			},
		},
	})
}

func TestAccLoginItemDataSourceByUUID(t *testing.T) {
	config, err := utils.GetTestConfig()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	vaultID := "t7dnwbjh6nlyw475wl3m442sdi"
	itemUUID := "dsrwv5dyacw4f7pdrfnmh36pne"

	itemTitle := "Test Login"
	username := "testUsername"
	password := "testPassword"


	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoginItemDataByUUIDConfig(config, itemUUID, vaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_item.test", "uuid", itemUUID),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", "login"),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "title", itemTitle),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "username", username),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "password", password),

				),
			},
		},
	})
}

func TestAccLoginItemDataSource_InvalidUUID(t *testing.T) {
    config, err := utils.GetTestConfig()
    if err != nil {
        t.Fatalf("Failed to get test config: %v", err)
    }

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccLoginItemDataByUUIDConfig(config, "invalid-uuid", "t7dnwbjh6nlyw475wl3m442sdi"),
                ExpectError: regexp.MustCompile(`"invalid-uuid" isn't an item`),
            },
        },
    })
}

func TestAccLoginItemDataSource_InvalidTitle(t *testing.T) {
    config, err := utils.GetTestConfig()
    if err != nil {
        t.Fatalf("Failed to get test config: %v", err)
    }

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccLoginItemDataSourceByTitleConfig(config, "invalid-name", "t7dnwbjh6nlyw475wl3m442sdi"),
                ExpectError: regexp.MustCompile(`"invalid-name" isn't an item`),
            },
        },
    })
}


func testAccLoginItemDataSourceByTitleConfig(config *utils.TestConfig, itemTitle string, vaultID string) string {
	return fmt.Sprintf(`
%s

# Test reading a pre-existing item by title
data "onepassword_item" "test" {
  title = "%s"
  vault = "%s"
}
`, utils.GetProviderConfig(config), itemTitle, vaultID)
}

func testAccLoginItemDataByUUIDConfig(config *utils.TestConfig, itemUUID string, vaultID string) string {
    return fmt.Sprintf(`
%s

# Test reading a pre-existing item by UUID
data "onepassword_item" "test" {
  uuid  = "%s"
  vault = "%s"
}`, utils.GetProviderConfig(config), itemUUID, vaultID)
}

