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
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEphemeralItemResource_ReadByUUID(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadByTitle(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServerWithTitleLookup(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemByTitleConfig(expectedItem.VaultID, expectedItem.Title),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadLoginItem(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadPasswordItem(t *testing.T) {
	expectedItem := generatePasswordItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadDatabaseItem(t *testing.T) {
	expectedItem := generateDatabaseItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadSecureNoteItem(t *testing.T) {
	expectedItem := generateSecureNoteItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadSSHKeyItem(t *testing.T) {
	expectedItem := generateSSHKeyItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadApiCredentialItem(t *testing.T) {
	expectedItem := generateApiCredentialItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_ReadDocumentItem(t *testing.T) {
	expectedItem := generateDocumentItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + testAccEphemeralItemConfig(expectedItem.VaultID, expectedItem.ID),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func TestAccEphemeralItemResource_UseInWriteOnlyPasswordField(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + fmt.Sprintf(`
ephemeral "onepassword_item" "source" {
  vault = "%s"
  uuid  = "%s"
}

resource "onepassword_item" "test" {
  vault             = "%s"
  title             = "Test Item"
  category          = "login"
  username          = "testuser"
  url               = "https://example.com"
  password_wo       = ephemeral.onepassword_item.source.password
  password_wo_version = 1
}
`, expectedItem.VaultID, expectedItem.ID, expectedItem.VaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test", "title", "Test Item"),
					resource.TestCheckResourceAttr("onepassword_item.test", "username", "testuser"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEphemeralItemResource_UseInWriteOnlyNoteValueField(t *testing.T) {
	expectedItem := generateSecureNoteItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + fmt.Sprintf(`
ephemeral "onepassword_item" "source" {
  vault = "%s"
  uuid  = "%s"
}

resource "onepassword_item" "test" {
  vault                = "%s"
  title                = "Test Secure Note"
  category             = "secure_note"
  note_value_wo        = ephemeral.onepassword_item.source.note_value
  note_value_wo_version = 1
}
`, expectedItem.VaultID, expectedItem.ID, expectedItem.VaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test", "title", "Test Secure Note"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEphemeralItemResource_ValidationError_MissingUUIDAndTitle(t *testing.T) {
	expectedItem := generateLoginItem()
	expectedVault := model.Vault{
		ID:          expectedItem.VaultID,
		Name:        "Name of the vault",
		Description: "This vault will be retrieved",
	}

	testServer := setupTestServer(expectedItem, expectedVault, t)
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + fmt.Sprintf(`
ephemeral "onepassword_item" "test" {
  vault = "%s"
}
`, expectedItem.VaultID),
				ExpectError: regexp.MustCompile("No attribute specified when one \\(and only one\\) of"),
			},
		},
	})
}

func TestAccEphemeralItemResource_ValidationError_ItemNotFound(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/v1/vaults/gs2jpwmahszwq25a7jiw45e4je/items/nonexistent" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + `
ephemeral "onepassword_item" "test" {
  vault = "gs2jpwmahszwq25a7jiw45e4je"
  uuid  = "nonexistent"
}`,
				ExpectError: regexp.MustCompile("Unable to read item"),
			},
		},
	})
}

func testAccEphemeralItemConfig(vault, uuid string) string {
	return fmt.Sprintf(`
ephemeral "onepassword_item" "test" {
  vault = "%s"
  uuid  = "%s"
}
`, vault, uuid)
}

func testAccEphemeralItemByTitleConfig(vault, title string) string {
	return fmt.Sprintf(`
ephemeral "onepassword_item" "test" {
  vault = "%s"
  title = "%s"
}
`, vault, title)
}

func setupTestServerWithTitleLookup(expectedItem *model.Item, expectedVault model.Vault, t *testing.T) *httptest.Server {
	connectItem, err := expectedItem.FromModelItemToConnect()
	if err != nil {
		t.Errorf("error converting item to connect item: %s", err)
	}

	itemBytes, err := json.Marshal(connectItem)
	if err != nil {
		t.Errorf("error marshaling item for testing: %s", err)
	}

	connectVault := expectedVault.ToConnectVault()
	vaultBytes, err := json.Marshal(connectVault)
	if err != nil {
		t.Errorf("error marshaling vault for testing: %s", err)
	}

	connectItemList := []*onepassword.Item{connectItem}
	itemListBytes, err := json.Marshal(connectItemList)
	if err != nil {
		t.Errorf("error marshaling itemlist for testing: %s", err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedItem.VaultID, expectedItem.ID) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if r.URL.String() == fmt.Sprintf("/v1/vaults/%s", expectedItem.VaultID) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(vaultBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if strings.Contains(r.URL.String(), "/v1/vaults/") && strings.Contains(r.URL.String(), "/items") && strings.Contains(r.URL.Query().Get("filter"), expectedItem.Title) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemListBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else {
				t.Errorf("Unexpected request: %s Consider adding this endpoint to the test server", r.URL.String())
			}
		}
	}))
}
