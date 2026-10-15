package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

const (
	testVaultID = "gs2jpwmahszwq25a7jiw45e4je"
	testItemID  = "rix6gwgpuyog4gqplegvrp3dbm"
)

var (
	// The timestamps Connect reports in a write response: they describe the
	// state of the item *before* the write.
	staleCreatedAt = time.Date(2024, 3, 1, 10, 30, 0, 0, time.UTC)
	staleUpdatedAt = time.Date(2024, 3, 1, 10, 30, 0, 0, time.UTC)

	// The timestamps Connect reports once the write has propagated.
	propagatedCreatedAt = time.Date(2024, 3, 1, 10, 30, 0, 0, time.UTC)
	propagatedUpdatedAt = time.Date(2024, 5, 2, 11, 45, 0, 0, time.UTC)
)

func connectItem(version int, createdAt, updatedAt time.Time) *onepassword.Item {
	return &onepassword.Item{
		ID:        testItemID,
		Title:     "test item",
		Vault:     onepassword.ItemVault{ID: testVaultID},
		Category:  onepassword.ItemCategory("LOGIN"),
		Version:   version,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// itemServer serves the write response for writeMethod and the propagated item
// on GET. A nil propagated item makes GET report the item as missing, which is
// how Connect behaves until the write has synced.
func itemServer(t *testing.T, writeMethod string, writeResponse, propagated *onepassword.Item) *httptest.Server {
	t.Helper()

	writeJSON := func(w http.ResponseWriter, item *onepassword.Item) {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			t.Errorf("error marshaling item for testing: %s", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(itemBytes); err != nil {
			t.Errorf("error writing body: %s", err)
		}
	}

	itemURL := fmt.Sprintf("/v1/vaults/%s/items/%s", testVaultID, testItemID)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == writeMethod:
			writeJSON(w, writeResponse)
		case r.Method == http.MethodGet && r.URL.Path == itemURL:
			if propagated == nil {
				w.WriteHeader(http.StatusNotFound)
				if _, err := w.Write([]byte(`{"status":404,"message":"item not found"}`)); err != nil {
					t.Errorf("error writing body: %s", err)
				}
				return
			}
			writeJSON(w, propagated)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
}

func TestCreateItemReportsPropagatedTimestamps(t *testing.T) {
	tests := map[string]struct {
		propagated        *onepassword.Item
		expectedCreatedAt time.Time
		expectedUpdatedAt time.Time
	}{
		"should report the timestamps of the propagated item": {
			propagated:        connectItem(1, propagatedCreatedAt, propagatedUpdatedAt),
			expectedCreatedAt: propagatedCreatedAt,
			expectedUpdatedAt: propagatedUpdatedAt,
		},
		"should fall back to the create response when the item does not propagate": {
			propagated:        nil,
			expectedCreatedAt: staleCreatedAt,
			expectedUpdatedAt: staleUpdatedAt,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			server := itemServer(t, http.MethodPost, connectItem(1, staleCreatedAt, staleUpdatedAt), test.propagated)
			defer server.Close()

			client := NewClient(server.URL, "test-token", Config{ProviderUserAgent: "test/1.0.0"})

			created, err := client.CreateItem(context.Background(), &model.Item{
				ID:       testItemID,
				VaultID:  testVaultID,
				Title:    "test item",
				Category: model.Login,
			}, testVaultID)
			if err != nil {
				t.Fatalf("CreateItem() error = %v", err)
			}

			if !created.CreatedAt.Equal(test.expectedCreatedAt) {
				t.Errorf("CreatedAt = %v, expected %v", created.CreatedAt, test.expectedCreatedAt)
			}
			if !created.UpdatedAt.Equal(test.expectedUpdatedAt) {
				t.Errorf("UpdatedAt = %v, expected %v", created.UpdatedAt, test.expectedUpdatedAt)
			}
		})
	}
}

func TestUpdateItemReportsPropagatedTimestamps(t *testing.T) {
	// Connect answers a PUT with the pre-update item, so the update response
	// carries version 1 and the previous `updatedAt`.
	updateResponse := connectItem(1, staleCreatedAt, staleUpdatedAt)
	propagated := connectItem(2, propagatedCreatedAt, propagatedUpdatedAt)

	server := itemServer(t, http.MethodPut, updateResponse, propagated)
	defer server.Close()

	client := NewClient(server.URL, "test-token", Config{ProviderUserAgent: "test/1.0.0"})

	updated, err := client.UpdateItem(context.Background(), &model.Item{
		ID:       testItemID,
		VaultID:  testVaultID,
		Title:    "test item",
		Category: model.Login,
	}, testVaultID)
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}

	if !updated.UpdatedAt.Equal(propagatedUpdatedAt) {
		t.Errorf("UpdatedAt = %v, expected the time of the update %v (got the previous change instead)",
			updated.UpdatedAt, propagatedUpdatedAt)
	}
	if !updated.CreatedAt.Equal(propagatedCreatedAt) {
		t.Errorf("CreatedAt = %v, expected %v", updated.CreatedAt, propagatedCreatedAt)
	}
}
