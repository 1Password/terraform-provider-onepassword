package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

// titleFilter matches the filter Connect uses to look an item up by title, as built by
// connect-sdk-go's GetItemsByTitle.
var titleFilter = regexp.MustCompile(`^title eq "(.*)"$`)

func setupTestServerMultipleItems(expectedItems []*model.Item, expectedVault model.Vault, t *testing.T) *httptest.Server {
	t.Helper()

	connectItemBytes := make(map[string][]byte, len(expectedItems))

	var connectItemList []*onepassword.Item
	connectItemsByTitle := make(map[string][]*onepassword.Item, len(expectedItems))

	for _, item := range expectedItems {
		connectItem, err := item.FromModelItemToConnect()
		if err != nil {
			t.Fatalf("error converting item to connect item: %s", err)
		}
		connectItemList = append(connectItemList, connectItem)
		connectItemsByTitle[item.Title] = append(connectItemsByTitle[item.Title], connectItem)

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

	writeJSON := func(w http.ResponseWriter, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(body); err != nil {
			t.Errorf("error writing body: %s", err)
		}
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Unexpected method: %s", r.Method)
			return
		}

		vaultPath := fmt.Sprintf("/v1/vaults/%s", expectedVault.ID)
		itemsPath := vaultPath + "/items"

		switch {
		case r.URL.Path == vaultPath:
			writeJSON(w, vaultBytes)

		case r.URL.Path == itemsPath:
			// Connect looks an item up by title by filtering the vault listing rather than by
			// path, so the filter has to be honoured here for the title branch to work at all.
			match := titleFilter.FindStringSubmatch(r.URL.Query().Get("filter"))
			if match == nil {
				writeJSON(w, itemListBytes)
				return
			}
			matchedBytes, err := json.Marshal(connectItemsByTitle[match[1]])
			if err != nil {
				t.Errorf("error marshaling filtered itemlist for testing: %s", err)
				return
			}
			writeJSON(w, matchedBytes)

		case strings.HasPrefix(r.URL.Path, itemsPath+"/"):
			itemID := strings.TrimPrefix(r.URL.Path, itemsPath+"/")
			itemBytes, ok := connectItemBytes[itemID]
			if !ok {
				t.Errorf("Unexpected item requested: %s", itemID)
				return
			}
			writeJSON(w, itemBytes)

		default:
			t.Errorf("Unexpected request: %s", r.URL.String())
		}
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
					// Items requested by UUID are keyed by that UUID, not by their title.
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".id", loginItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".title", loginItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".category", strings.ToLower(string(loginItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".username", loginItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".password", loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+passwordItem.ID+".id", passwordItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+passwordItem.ID+".title", passwordItem.Title),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+passwordItem.ID+".category", strings.ToLower(string(passwordItem.Category))),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+loginItem.ID, loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+passwordItem.ID, passwordItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemsDataSourceByTitle(t *testing.T) {
	loginItem := generateLoginItem()
	loginItem.ID = "abcdefghijklmnopqrstuvwx10"
	loginItem.Title = "Login Secret"

	passwordItem := generatePasswordItem()
	passwordItem.ID = "abcdefghijklmnopqrstuvwx11"
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
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, loginItem.Title, passwordItem.Title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.Title+".id", loginItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+passwordItem.Title+".id", passwordItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+loginItem.Title, loginItem.Fields[1].Value),
				),
			},
		},
	})
}

// Two different items can share a title. Keying the output maps by the requested identifier is
// what keeps them apart, so each must come back with its own id and its own secret.
func TestAccItemsDataSourceSharedTitle(t *testing.T) {
	firstItem := generateLoginItem()
	firstItem.ID = "abcdefghijklmnopqrstuvwx20"
	firstItem.Title = "Shared Title"
	firstItem.Fields[1].Value = "first_password"

	secondItem := generateLoginItem()
	secondItem.ID = "abcdefghijklmnopqrstuvwx21"
	secondItem.Title = "Shared Title"
	secondItem.Fields[1].Value = "second_password"

	expectedVault := model.Vault{
		ID:          firstItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerMultipleItems([]*model.Item{firstItem, secondItem}, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, firstItem.ID, secondItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items.%", "2"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+firstItem.ID+".id", firstItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+secondItem.ID+".id", secondItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+firstItem.ID, "first_password"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+secondItem.ID, "second_password"),
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
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".category", "login"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+apiCredItem.ID+".category", "api_credential"),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+passwordItem.ID+".category", "password"),
					// credentials map: API credential uses credential field, others use password
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+loginItem.ID, loginItem.Fields[1].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+apiCredItem.ID, apiCredItem.Fields[1].Value), // credential field
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+passwordItem.ID, passwordItem.Fields[1].Value),
				),
			},
		},
	})
}

// An item with neither a credential nor a password has no entry in the credentials map, and its
// absent fields are null rather than empty strings.
func TestAccItemsDataSourceItemWithoutCredential(t *testing.T) {
	noteItem := generateSecureNoteItem()
	noteItem.ID = "abcdefghijklmnopqrstuvwx30"
	noteItem.Title = "My Note"

	expectedVault := model.Vault{
		ID:          noteItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerMultipleItems([]*model.Item{noteItem}, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, noteItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+noteItem.ID+".note_value", noteItem.Fields[0].Value),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials.%", "0"),
					resource.TestCheckNoResourceAttr("data.onepassword_items.test", "credentials."+noteItem.ID),
					resource.TestCheckNoResourceAttr("data.onepassword_items.test", "items."+noteItem.ID+".credential"),
					resource.TestCheckNoResourceAttr("data.onepassword_items.test", "items."+noteItem.ID+".password"),
					resource.TestCheckNoResourceAttr("data.onepassword_items.test", "items."+noteItem.ID+".url"),
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
				Config: testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, loginItem.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.onepassword_items.test", "items."+loginItem.ID+".id", loginItem.ID),
					resource.TestCheckResourceAttr("data.onepassword_items.test", "credentials."+loginItem.ID, loginItem.Fields[1].Value),
				),
			},
		},
	})
}

func TestAccItemsDataSourceDuplicateIdentifiers(t *testing.T) {
	loginItem := generateLoginItem()
	loginItem.ID = "abcdefghijklmnopqrstuvwx40"

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
				Config:      testAccProviderConfig(testServer.URL) + testAccItemsDataSourceConfig(expectedVault.ID, loginItem.ID, loginItem.ID),
				ExpectError: regexp.MustCompile(`(?s)Duplicate List Value`),
			},
		},
	})
}

func testAccItemsDataSourceConfig(vault string, titlesOrIDs ...string) string {
	quoted := make([]string, len(titlesOrIDs))
	for i, t := range titlesOrIDs {
		quoted[i] = fmt.Sprintf("%q", t)
	}
	return fmt.Sprintf(`
data "onepassword_items" "test" {
  vault  = "%s"
  titles = [%s]
}`, vault, strings.Join(quoted, ", "))
}
