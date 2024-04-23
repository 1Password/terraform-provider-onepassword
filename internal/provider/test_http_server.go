package provider

import (
	"encoding/json"
	"fmt"
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

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == fmt.Sprintf("/v1/vaults/%s/items/%s", expectedItem.Vault.ID, expectedItem.ID) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(itemBytes)
			if err != nil {
				t.Errorf("error writing body: %s", err)
			}
		} else if r.URL.Path == fmt.Sprintf("/v1/vaults/%s", expectedItem.Vault.ID) {
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
