package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

// setupTestServer sets up a http server that can be used mock out 1Password Connect API calls
func setupTestServer(expectedItem *model.Item, expectedVault model.Vault, t *testing.T) *httptest.Server {
	connectItem, err := expectedItem.FromModelItemToConnect()
	if err != nil {
		t.Errorf("error converting item to connect item: %s", err)
	}

	itemBytes, err := json.Marshal(connectItem)
	if err != nil {
		t.Errorf("error marshaling item for testing: %s", err)
	}

	files := expectedItem.Files
	var fileBytes [][]byte
	for _, file := range files {
		c, err := file.Content()
		if err != nil {
			t.Errorf("error getting file content: %s", err)
		}
		fileBytes = append(fileBytes, c)
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
		filePath := regexp.MustCompile("/v1/vaults/[a-z0-9]*/items/[a-z0-9]*/files/[a-z0-9]*/content")
		if r.Method == http.MethodGet {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedItem.VaultID, expectedItem.ID) {
				// Mock returning an item specified by uuid
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if r.URL.String() == fmt.Sprintf("/v1/vaults/%s", expectedItem.VaultID) {
				// Mock returning a vault specified by uuid
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(vaultBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items", expectedItem.VaultID) {
				// Mock returning a list of items for a vault specified by uuid
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemListBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if filePath.MatchString(r.URL.String()) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("1Password-Connect-Version", "1.3.0") // must be >= 1.3.0
				i := slices.IndexFunc(files, func(f model.ItemFile) bool {
					return f.ID == strings.Split(r.URL.Path, "/")[7]
				})
				if i == -1 {
					t.Errorf("file not found")
				}
				_, err := w.Write(fileBytes[i])
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else {
				t.Errorf("Unexpected request: %s Consider adding this endpoint to the test server", r.URL.String())
			}
		} else if r.Method == http.MethodPost {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items", expectedItem.VaultID) {
				itemToReturn := convertBodyToItem(r, t)

				if itemToReturn.Category != model.SecureNote {
					itemField := model.ItemField{
						Label: "password",
						Value: "somepassword",
					}
					itemToReturn.Fields = append(itemToReturn.Fields, itemField)
				}

				// Set the ID and VaultID
				itemToReturn.ID = expectedItem.ID
				itemToReturn.VaultID = expectedItem.VaultID

				// Convert back to Connect format for response
				connectItemToReturn, err := itemToReturn.FromModelItemToConnect()
				if err != nil {
					t.Errorf("error converting model item to Connect format: %s", err)
				}
				itemBytes, err := json.Marshal(connectItemToReturn)

				if err != nil {
					t.Errorf("error marshaling item for testing: %s", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, err = w.Write(itemBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else {
				t.Errorf("Unexpected request: %s Consider adding this endpoint to the test server", r.URL.String())
			}
		} else if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
		} else {
			t.Errorf("Method not supported: %s", r.Method)
		}
	}))
}

func convertBodyToItem(r *http.Request, t *testing.T) model.Item {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading item body for testing: %s", err)
	}
	connectItemToReturn := onepassword.Item{}
	err = json.Unmarshal(rawBody, &connectItemToReturn)
	if err != nil {
		t.Errorf("error unmarshaling item for testing: %s", err)
	}

	modelItemToReturn := model.Item{}
	err = modelItemToReturn.FromConnectItemToModel(&connectItemToReturn)
	if err != nil {
		t.Errorf("error converting Connect item to model: %s", err)
	}

	return modelItemToReturn
}
