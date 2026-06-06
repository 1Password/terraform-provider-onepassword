package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func setupTestServerMultipleItems(expectedItems []*model.Item, expectedVault model.Vault, t *testing.T) *httptest.Server {
	t.Helper()

	connectItems := make(map[string]*onepassword.Item, len(expectedItems))
	connectItemBytes := make(map[string][]byte, len(expectedItems))

	var connectItemList []*onepassword.Item

	for _, item := range expectedItems {
		connectItem, err := item.FromModelItemToConnect()
		if err != nil {
			t.Fatalf("error converting item to connect item: %s", err)
		}
		connectItems[item.ID] = connectItem
		connectItemList = append(connectItemList, connectItem)

		itemBytes, err := json.Marshal(connectItem)
		if err != nil {
			t.Fatalf("error marshaling item for testing: %s", err)
		}
		connectItemBytes[item.ID] = itemBytes
	}

	connectVault := expectedVault.ToConnectVault()
	vaultBytes, err := json.Marshal(connectVault)
	if err != nil {
		t.Fatalf("error marshaling vault for testing: %s", err)
	}

	itemListBytes, err := json.Marshal(connectItemList)
	if err != nil {
		t.Fatalf("error marshaling itemlist for testing: %s", err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Unexpected method: %s", r.Method)
			return
		}

		// Match /v1/vaults/{vaultID}/items/{itemID}
		for _, item := range expectedItems {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedVault.ID, item.ID) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(connectItemBytes[item.ID])
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
				return
			}
		}

		if r.URL.String() == fmt.Sprintf("/v1/vaults/%s", expectedVault.ID) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(vaultBytes)
			if err != nil {
				t.Errorf("error writing body: %s", err)
			}
			return
		}

		if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items", expectedVault.ID) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(itemListBytes)
			if err != nil {
				t.Errorf("error writing body: %s", err)
			}
			return
		}

		t.Errorf("Unexpected request: %s", r.URL.String())
	}))
}

func TestAccItemsDataSourceBatchRead(t *testing.T) {
	loginItem := generateLoginItem()
	loginItem.ID = "abcdefghijklmnopqrstuvwx01"
	loginItem.Title = "Login Secret"

	passwordItem := generatePasswordItem()
	passwordItem.ID = "abcdefghijklmnopqrstuvwx02"
	passwordItem.Title = "Password Secret"

	expectedVault := model.Vault{
		ID:          loginItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerMultipleItems([]*model.Item{loginItem, passwordItem}, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, loginItem.ID, passwordItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "vault", expectedVault.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Login Secret.id", loginItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Login Secret.title", loginItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Login Secret.category", strings.ToLower(string(loginItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Login Secret.username", loginItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Login Secret.password", loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Password Secret.id", passwordItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Password Secret.title", passwordItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Password Secret.category", strings.ToLower(string(passwordItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.Login Secret", loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.Password Secret", passwordItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemsDataSourceMixedCategories(t *testing.T) {
	loginItem := generateLoginItem()
	loginItem.ID = "abcdefghijklmnopqrstuvwx03"
	loginItem.Title = "My Login"

	apiCredItem := generateApiCredentialItem()
	apiCredItem.ID = "abcdefghijklmnopqrstuvwx04"
	apiCredItem.Title = "My API Key"

	passwordItem := generatePasswordItem()
	passwordItem.ID = "abcdefghijklmnopqrstuvwx05"
	passwordItem.Title = "My Password"

	expectedVault := model.Vault{
		ID:          loginItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerMultipleItems([]*model.Item{loginItem, apiCredItem, passwordItem}, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, loginItem.ID, apiCredItem.ID, passwordItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.My Login.category", "login"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.My API Key.category", "api_credential"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.My Password.category", "password"),
					// credentials map: API credential uses credential field, others use password
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.My Login", loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.My API Key", apiCredItem.Fields[1].Value), // credential field
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.My Password", passwordItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemsDataSourceSingleItem(t *testing.T) {
	loginItem := generateLoginItem()
	loginItem.ID = "abcdefghijklmnopqrstuvwx06"
	loginItem.Title = "Solo Item"

	expectedVault := model.Vault{
		ID:          loginItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerMultipleItems([]*model.Item{loginItem}, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + fmt.Sprintf(`
data "onepassword_items" "test" {
  vault  = "%s"
  titles = ["%s"]
}`, expectedVault.ID, loginItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.Solo Item.id", loginItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.Solo Item", loginItem.Fields[1].Value),
				),
			},
		},
	})
}

func testAccItemsDataSourceConfig(vault string, uuids ...string) string {
	quoted := make([]string, len(uuids))
	for i, u := range uuids {
		quoted[i] = fmt.Sprintf("%q", u)
	}
	return fmt.Sprintf(`
data "onepassword_items" "test" {
  vault  = "%s"
  titles = [%s]
}`, vault, strings.Join(quoted, ", "))
}
