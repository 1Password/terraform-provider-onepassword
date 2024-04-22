package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccItemDataSource(t *testing.T) {
	expectedItem := generateItem()
	expectedVault := onepassword.Vault{
		ID:          expectedItem.Vault.ID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccItemDataSourceConfig(expectedItem.Vault.ID, expectedItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_item.test", "id", fmt.Sprintf("vaults/%s/items/%s", expectedVault.ID, expectedItem.ID)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "title", expectedItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "category", string(expectedItem.Category)),
					resource.TestCheckResourceAttr("data.onepassword_item.test", "urls", string(expectedItem.URLs[0].URL)),
				),
			},
		},
	})
}

func setupTestServer(expectedItem *onepassword.Item, expectedVault onepassword.Vault, t *testing.T) *httptest.Server {
	itemBytes, err := json.Marshal(expectedItem)
	if err != nil {
		t.Errorf("error marshaling item for testing: %s", err)
	}

	vaultBytes, err := json.Marshal(expectedVault)
	if err != nil {
		t.Errorf("error marshaling vault for testing: %s", err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedItem.Vault.ID, expectedItem.ID) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(itemBytes)
			if err != nil {
				t.Errorf("error writing body: %s", err)
			}
		} else if r.URL.Path == fmt.Sprintf("/v1/vaults/%s", expectedItem.Vault.ID) {
			t.Errorf("specific path: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(vaultBytes)
			if err != nil {
				t.Errorf("error writing body: %s", err)
			}
		} else {
			t.Errorf("not handled")
		}
	}))
}

func generateItem() *onepassword.Item {
	item := onepassword.Item{}
	item.Fields = generateFields()
	item.ID = "rix6gwgpuyog4gqplegvrp3dbm"
	item.Vault.ID = "gs2jpwmahszwq25a7jiw45e4je"
	item.Category = "CUSTOM"
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
