package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
)

// setupTestServer sets up a http server that can be used mock out 1Password Connect API calls
func setupTestServer(expectedItem *onepassword.Item, expectedVault onepassword.Vault, t *testing.T) *httptest.Server {
	itemBytes, err := json.Marshal(expectedItem)
	if err != nil {
		t.Errorf("error marshaling item for testing: %s", err)
	}

	vaultBytes, err := json.Marshal(expectedVault)
	if err != nil {
		t.Errorf("error marshaling vault for testing: %s", err)
	}

	itemList := []onepassword.Item{*expectedItem}
	itemListBytes, err := json.Marshal(itemList)
	if err != nil {
		t.Errorf("error marshaling itemlist for testing: %s", err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t.Errorf("the url %s and the method: %s", r.URL.String(), r.Method)
		if r.Method == http.MethodGet {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedItem.Vault.ID, expectedItem.ID) {
				// Mock returning an item specified by uuid
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if r.URL.String() == fmt.Sprintf("/v1/vaults/%s", expectedItem.Vault.ID) {
				// Mock returning a vault specified by uuid
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(vaultBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items", expectedItem.Vault.ID) {
				// Mock returning a list of items for a vault specified by uuid
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(itemListBytes)
				if err != nil {
					t.Errorf("error writing body: %s", err)
				}
			} else {
				t.Errorf("Unexpected request: %s Consider adding this endpoint to the test server", r.URL.String())
			}
		} else if r.Method == http.MethodPost {
			if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items", expectedItem.Vault.ID) {
				itemToReturn := convertBodyToItem(r, t)
				itemToReturn.Fields = []*onepassword.ItemField{
					{
						Label: "password",
						Value: "somepassword",
					},
				}
				itemToReturn.ID = expectedItem.ID
				itemBytes, err := json.Marshal(itemToReturn)
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

func convertBodyToItem(r *http.Request, t *testing.T) onepassword.Item {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading item body for testing: %s", err)
	}
	itemToReturn := onepassword.Item{}
	err = json.Unmarshal(rawBody, &itemToReturn)
	if err != nil {
		t.Errorf("error unmarshaling item for testing: %s", err)
	}

	return itemToReturn

}
