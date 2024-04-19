// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// func TestDataSourceOnePasswordItemRead(t *testing.T) {
// 	expectedItem := generateItem()
// 	&testClient{
// 		GetItemFunc: func(uuid string, vaultUUID string) (*onepassword.Item, error) {
// 			return expectedItem, nil
// 		},
// 	}

// 	dataSourceData := generateDataSource(t, expectedItem)

// }

// func generateDataSource(t *testing.T, item *onepassword.Item) *schema.ResourceData {
// 	dataSourceData := schema.TestResourceDataRaw(t, dataSourceOnepasswordItem().Schema, nil)
// 	dataSourceData.Set("vault", item.Vault.ID)
// 	dataSourceData.SetId(fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID))
// 	return dataSourceData
// }

// func generateItem() *onepassword.Item {
// 	item := onepassword.Item{}
// 	item.Fields = generateFields()
// 	item.ID = "79841a98-dd4a-4c34-8be5-32dca20a7328"
// 	item.Vault.ID = "df2e9643-45ad-4ff9-8b98-996f801afa75"
// 	item.Category = "USERNAME"
// 	item.Title = "test item"
// 	item.URLs = []onepassword.ItemURL{
// 		{
// 			Primary: true,
// 			URL:     "some_url.com",
// 		},
// 	}
// 	return &item
// }

// func generateFields() []*onepassword.ItemField {
// 	fields := []*onepassword.ItemField{
// 		{
// 			Label: "username",
// 			Value: "test_user",
// 		},
// 		{
// 			Label: "password",
// 			Value: "test_password",
// 		},
// 		{
// 			Label: "hostname",
// 			Value: "test_host",
// 		},
// 		{
// 			Label: "database",
// 			Value: "test_database",
// 		},
// 		{
// 			Label: "port",
// 			Value: "test_port",
// 		},
// 		{
// 			Label: "type",
// 			Value: "test_type",
// 		},
// 	}
// 	return fields
// }

func TestAccItemDataSource(t *testing.T) {
	expectedItem := generateItem()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccItemDataSourceConfig(expectedItem.Vault.ID, expectedItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					//testDataSourceOnePasswordItemRead(expectedItem),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "vault", expectedItem.Vault.ID),
					// testDataSourceOnePasswordItemRead()
				),
			},
		},
	})
}

func testDataSourceOnePasswordItemRead(expectedItem *onepassword.Item) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return fmt.Errorf("some error")
	}

	// expectedItem := generateItem()
	// meta := &testClient{
	// 	GetItemFunc: func(uuid string, vaultUUID string) (*onepassword.Item, error) {
	// 		return expectedItem, nil
	// 	},
	// }

	// dataSourceData := generateDataSource(t, expectedItem)
	// dataSourceData.Set("uuid", expectedItem.ID)

	// err := dataSourceOnepasswordItemRead(context.Background(), dataSourceData, meta)
	// if err != nil {
	// 	t.Errorf("Unexpected error occured")
	// }
	// compareItemToSource(t, dataSourceData, expectedItem)
}

func generateItem() *onepassword.Item {
	item := onepassword.Item{}
	item.Fields = generateFields()
	item.ID = "79841a98-dd4a-4c34-8be5-32dca20a7328"
	item.Vault.ID = "df2e9643-45ad-4ff9-8b98-996f801afa75"
	item.Category = "USERNAME"
	item.Title = "test item"
	item.URLs = []onepassword.ItemURL{
		{
			Primary: true,
			URL:     "some_url.com",
		},
	}
	return &item
}

func generateFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test_user",
		},
		{
			Label: "password",
			Value: "test_password",
		},
		{
			Label: "hostname",
			Value: "test_host",
		},
		{
			Label: "database",
			Value: "test_database",
		},
		{
			Label: "port",
			Value: "test_port",
		},
		{
			Label: "type",
			Value: "test_type",
		},
	}
	return fields
}
func testAccItemDataSourceConfig(vault, uuid string) string {
	return fmt.Sprintf(`
data "onepassword_item" "test" {
  vault = "%s"
  uuid = "%s"
}`, vault, uuid)
}
