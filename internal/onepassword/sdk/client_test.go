package sdk

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/1password/onepassword-sdk-go"
)

const (
	testVaultID = "gs2jpwmahszwq25a7jiw45e4je"
	alphaID     = "aaaaaaaaaaaaaaaaaaaaaaaaaa"
	betaID      = "bbbbbbbbbbbbbbbbbbbbbbbbbb"
	gammaID     = "cccccccccccccccccccccccccc"
)

// fakeItemsAPI embeds sdk.ItemsAPI so that any method GetItems does not use panics rather than
// silently returning a zero value.
type fakeItemsAPI struct {
	sdk.ItemsAPI

	overviews []sdk.ItemOverview
	getAll    func(itemIDs []string) (sdk.ItemsGetAllResponse, error)

	listCalls int
	getAllIDs []string
}

func (f *fakeItemsAPI) List(_ context.Context, _ string, _ ...sdk.ItemListFilter) ([]sdk.ItemOverview, error) {
	f.listCalls++
	return f.overviews, nil
}

func (f *fakeItemsAPI) GetAll(_ context.Context, _ string, itemIDs []string) (sdk.ItemsGetAllResponse, error) {
	f.getAllIDs = itemIDs
	return f.getAll(itemIDs)
}

func overview(id, title string) sdk.ItemOverview {
	return sdk.ItemOverview{ID: id, Title: title, VaultID: testVaultID, Category: sdk.ItemCategoryLogin}
}

// respond returns the requested items in the order they were asked for, which is what a
// well-behaved batch response looks like.
func respond(byID map[string]string) func([]string) (sdk.ItemsGetAllResponse, error) {
	return func(itemIDs []string) (sdk.ItemsGetAllResponse, error) {
		var response sdk.ItemsGetAllResponse
		for _, id := range itemIDs {
			item := sdk.Item{ID: id, Title: byID[id], VaultID: testVaultID, Category: sdk.ItemCategoryLogin}
			response.IndividualResponses = append(response.IndividualResponses, sdk.Response[sdk.Item, sdk.ItemsGetAllError]{Content: &item})
		}
		return response, nil
	}
}

func TestGetItems(t *testing.T) {
	titlesByID := map[string]string{alphaID: "Alpha", betaID: "Beta", gammaID: "Gamma"}
	allOverviews := []sdk.ItemOverview{overview(alphaID, "Alpha"), overview(betaID, "Beta"), overview(gammaID, "Gamma")}

	tests := map[string]struct {
		overviews     []sdk.ItemOverview
		getAll        func(itemIDs []string) (sdk.ItemsGetAllResponse, error)
		request       []string
		wantTitles    []string
		wantGetAllIDs []string
		wantListCalls int
		wantErr       string
	}{
		"ids alone are fetched without listing the vault": {
			getAll:        respond(titlesByID),
			request:       []string{alphaID, betaID},
			wantTitles:    []string{"Alpha", "Beta"},
			wantGetAllIDs: []string{alphaID, betaID},
			wantListCalls: 0,
		},
		"titles are resolved to ids before fetching": {
			overviews:     allOverviews,
			getAll:        respond(titlesByID),
			request:       []string{"Beta", "Alpha"},
			wantTitles:    []string{"Beta", "Alpha"},
			wantGetAllIDs: []string{betaID, alphaID},
			wantListCalls: 1,
		},
		"mixed titles and ids come back in request order": {
			overviews:     allOverviews,
			getAll:        respond(titlesByID),
			request:       []string{"Alpha", betaID, "Gamma"},
			wantTitles:    []string{"Alpha", "Beta", "Gamma"},
			wantGetAllIDs: []string{betaID, alphaID, gammaID},
			wantListCalls: 1,
		},
		"one item requested by both id and title appears twice": {
			overviews:     allOverviews,
			getAll:        respond(titlesByID),
			request:       []string{alphaID, "Alpha"},
			wantTitles:    []string{"Alpha", "Alpha"},
			wantListCalls: 1,
		},
		"a title matching no item is an error": {
			overviews: allOverviews,
			getAll:    respond(titlesByID),
			request:   []string{"Delta"},
			wantErr:   `found 0 item(s) in vault "gs2jpwmahszwq25a7jiw45e4je" with title "Delta"`,
		},
		"a title matching several items is an error": {
			overviews: append(allOverviews, overview(gammaID, "Alpha")),
			getAll:    respond(titlesByID),
			request:   []string{"Alpha"},
			wantErr:   `found 2 item(s) in vault "gs2jpwmahszwq25a7jiw45e4je" with title "Alpha"`,
		},
		"a failed response names the request position": {
			getAll: func([]string) (sdk.ItemsGetAllResponse, error) {
				notFound := sdk.NewItemsGetAllErrorTypeVariantItemNotFound()
				return sdk.ItemsGetAllResponse{IndividualResponses: []sdk.Response[sdk.Item, sdk.ItemsGetAllError]{
					{Content: &sdk.Item{ID: alphaID, Title: "Alpha", VaultID: testVaultID}},
					{Error: &notFound},
				}}, nil
			},
			request: []string{alphaID, betaID},
			wantErr: `request index 1 (likely "bbbbbbbbbbbbbbbbbbbbbbbbbb"): itemNotFound`,
		},
		"a response shorter than the request is an error": {
			getAll: func([]string) (sdk.ItemsGetAllResponse, error) {
				return sdk.ItemsGetAllResponse{IndividualResponses: []sdk.Response[sdk.Item, sdk.ItemsGetAllError]{
					{Content: &sdk.Item{ID: alphaID, Title: "Alpha", VaultID: testVaultID}},
				}}, nil
			},
			request: []string{alphaID, betaID},
			wantErr: `item "bbbbbbbbbbbbbbbbbbbbbbbbbb" not found in batch response`,
		},
		"a response longer than the request is an error": {
			getAll: func([]string) (sdk.ItemsGetAllResponse, error) {
				return sdk.ItemsGetAllResponse{IndividualResponses: []sdk.Response[sdk.Item, sdk.ItemsGetAllError]{
					{Content: &sdk.Item{ID: alphaID, Title: "Alpha", VaultID: testVaultID}},
					{Content: &sdk.Item{ID: betaID, Title: "Beta", VaultID: testVaultID}},
				}}, nil
			},
			request: []string{alphaID},
			wantErr: "batch get returned 2 responses for 1 requested items",
		},
		"an item nobody asked for does not satisfy the request": {
			getAll: func([]string) (sdk.ItemsGetAllResponse, error) {
				return sdk.ItemsGetAllResponse{IndividualResponses: []sdk.Response[sdk.Item, sdk.ItemsGetAllError]{
					{Content: &sdk.Item{ID: gammaID, Title: "Gamma", VaultID: testVaultID}},
				}}, nil
			},
			request: []string{alphaID},
			wantErr: `item "aaaaaaaaaaaaaaaaaaaaaaaaaa" not found in batch response`,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			items := &fakeItemsAPI{overviews: test.overviews, getAll: test.getAll}
			client := &Client{sdkClient: &sdk.Client{ItemsAPI: items}}

			got, err := client.GetItems(context.Background(), testVaultID, test.request)

			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", test.wantErr)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Errorf("Expected error containing %q, got %q", test.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %s", err)
			}

			gotTitles := make([]string, len(got))
			for i, item := range got {
				if item == nil {
					t.Fatalf("Expected an item at index %d, got nil", i)
				}
				gotTitles[i] = item.Title
			}
			if !equalStrings(gotTitles, test.wantTitles) {
				t.Errorf("Expected titles %+v, got %+v", test.wantTitles, gotTitles)
			}

			if test.wantGetAllIDs != nil && !equalStrings(items.getAllIDs, test.wantGetAllIDs) {
				t.Errorf("Expected batch get of %+v, got %+v", test.wantGetAllIDs, items.getAllIDs)
			}

			if items.listCalls != test.wantListCalls {
				t.Errorf("Expected %d vault listing(s), got %d", test.wantListCalls, items.listCalls)
			}
		})
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
