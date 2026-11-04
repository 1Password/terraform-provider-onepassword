package provider

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	sdk "github.com/1password/onepassword-sdk-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	opsdk "github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/sdk"
)

const testDocumentVaultID = "gs2jpwmahszwq25a7jiw45e4je"

// fakeSDKItems is an in-memory implementation of the parts of the 1Password SDK
// items API that the item resource uses, so document items can be tested through
// the real SDK client code path without a 1Password account.
type fakeSDKItems struct {
	sdk.ItemsAPI // methods the resource doesn't use panic if called

	mu       sync.Mutex
	items    map[string]sdk.Item
	files    map[string][]byte // document content by file ID
	nextID   int
	replaced int // number of ReplaceDocument calls
}

type fakeSDKFiles struct {
	sdk.ItemsFilesAPI
	items *fakeSDKItems
}

func newFakeSDKItems() *fakeSDKItems {
	return &fakeSDKItems{items: map[string]sdk.Item{}, files: map[string][]byte{}}
}

// id returns a fake 1Password UUID (26 lowercase alphanumeric characters).
func (f *fakeSDKItems) id(prefix string) string {
	f.nextID++
	return fmt.Sprintf("%s%024d", prefix, f.nextID)
}

func (f *fakeSDKItems) storeDocument(params *sdk.DocumentCreateParams) *sdk.FileAttributes {
	fileID := f.id("fd")
	f.files[fileID] = append([]byte(nil), params.Content...)
	return &sdk.FileAttributes{ID: fileID, Name: params.Name, Size: uint32(len(params.Content))}
}

func (f *fakeSDKItems) Create(_ context.Context, params sdk.ItemCreateParams) (sdk.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC().Truncate(time.Second)
	item := sdk.Item{
		ID:        f.id("it"),
		Title:     params.Title,
		Category:  params.Category,
		VaultID:   params.VaultID,
		Fields:    params.Fields,
		Sections:  params.Sections,
		Tags:      params.Tags,
		Websites:  params.Websites,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if params.Notes != nil {
		item.Notes = *params.Notes
	}
	if params.Document != nil {
		item.Document = f.storeDocument(params.Document)
	}
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeSDKItems) Get(_ context.Context, vaultID, itemID string) (sdk.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.items[itemID]
	if !ok || item.VaultID != vaultID {
		return sdk.Item{}, errors.New("item couldn't be found")
	}
	return item, nil
}

func (f *fakeSDKItems) Put(_ context.Context, item sdk.Item) (sdk.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	current, ok := f.items[item.ID]
	if !ok {
		return sdk.Item{}, errors.New("item couldn't be found")
	}
	// Like the real API, the document file can't be changed with Put.
	item.Document = current.Document
	item.Version = current.Version + 1
	item.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeSDKItems) Delete(_ context.Context, _ string, itemID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.items, itemID)
	return nil
}

func (f *fakeSDKItems) Files() sdk.ItemsFilesAPI {
	return fakeSDKFiles{items: f}
}

func (ff fakeSDKFiles) Read(_ context.Context, _ string, itemID string, attr sdk.FileAttributes) ([]byte, error) {
	f := ff.items
	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.items[itemID]
	if !ok || item.Document == nil || item.Document.ID != attr.ID {
		return nil, fmt.Errorf("file %q not found in item %q", attr.ID, itemID)
	}
	return append([]byte(nil), f.files[attr.ID]...), nil
}

func (ff fakeSDKFiles) ReplaceDocument(_ context.Context, item sdk.Item, params sdk.DocumentCreateParams) (sdk.Item, error) {
	f := ff.items
	f.mu.Lock()
	defer f.mu.Unlock()

	current, ok := f.items[item.ID]
	if !ok {
		return sdk.Item{}, errors.New("item couldn't be found")
	}
	if current.Category != sdk.ItemCategoryDocument {
		return sdk.Item{}, errors.New("item is not a document")
	}
	f.replaced++
	current.Document = f.storeDocument(&params)
	current.Version++
	f.items[item.ID] = current
	return current, nil
}

// only returns the single stored item, failing the test if there isn't exactly one.
func (f *fakeSDKItems) only(t *testing.T) (sdk.Item, []byte) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.items) != 1 {
		t.Fatalf("expected exactly 1 item in the fake vault, found %d", len(f.items))
	}
	for _, item := range f.items {
		if item.Document == nil {
			return item, nil
		}
		return item, f.files[item.Document.ID]
	}
	return sdk.Item{}, nil
}

// setDocumentContent changes the stored document file as if it was edited outside Terraform.
func (f *fakeSDKItems) setDocumentContent(t *testing.T, content []byte) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, item := range f.items {
		if item.Document != nil {
			f.files[item.Document.ID] = content
			return
		}
	}
	t.Fatal("no document item in the fake vault")
}

// useFakeSDK makes the provider talk to the given fake SDK backend for the rest of the test.
func useFakeSDK(t *testing.T, items *fakeSDKItems) {
	t.Helper()
	original := newClient
	newClient = func(context.Context, onepassword.ClientConfig) (onepassword.Client, error) {
		return opsdk.NewClientFromSDK(&sdk.Client{ItemsAPI: items}), nil
	}
	t.Cleanup(func() { newClient = original })
}

const testServiceAccountProviderConfig = `
provider "onepassword" {
  service_account_token = "ops_test"
}
`

func testAccDocumentConfig(title, document string) string {
	return testServiceAccountProviderConfig + fmt.Sprintf(`
resource "onepassword_item" "doc" {
  vault      = %q
  title      = %q
  category   = "document"
  note_value = "rotated yearly"
  tags       = ["tls"]

  document = {
%s
  }
}
`, testDocumentVaultID, title, document)
}

// checkDocumentStored verifies what the fake vault holds after an apply.
func checkDocumentStored(items *fakeSDKItems, t *testing.T, title, name string, content []byte) resource.TestCheckFunc {
	return func(*terraform.State) error {
		item, stored := items.only(t)
		if item.Category != sdk.ItemCategoryDocument {
			return fmt.Errorf("expected category %q, got %q", sdk.ItemCategoryDocument, item.Category)
		}
		if item.Title != title {
			return fmt.Errorf("expected title %q, got %q", title, item.Title)
		}
		if item.Notes != "rotated yearly" {
			return fmt.Errorf("expected notes to be saved, got %q", item.Notes)
		}
		if item.Document == nil || item.Document.Name != name {
			return fmt.Errorf("expected document named %q, got %+v", name, item.Document)
		}
		if string(stored) != string(content) {
			return fmt.Errorf("expected document content %q, got %q", content, stored)
		}
		return nil
	}
}

func checkReplaced(items *fakeSDKItems, expected int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		items.mu.Lock()
		defer items.mu.Unlock()
		if items.replaced != expected {
			return fmt.Errorf("expected %d document uploads after create, got %d", expected, items.replaced)
		}
		return nil
	}
}

// captureAttr stores an attribute value from state; sameAttr checks it later.
func captureAttr(name, key string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found", name)
		}
		*into = rs.Primary.Attributes[key]
		return nil
	}
}

func sameAttr(name, key string, previous *string, wantSame bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found", name)
		}
		current := rs.Primary.Attributes[key]
		if wantSame && current != *previous {
			return fmt.Errorf("expected %s to stay %q, got %q", key, *previous, current)
		}
		if !wantSame && current == *previous {
			return fmt.Errorf("expected %s to change from %q", key, *previous)
		}
		return nil
	}
}

func TestAccItemResourceDocument_Lifecycle(t *testing.T) {
	items := newFakeSDKItems()
	useFakeSDK(t, items)

	binary := []byte{0x00, 0xff, 0x10, 0x80, 'p', '1', '2'}
	binaryBase64 := base64.StdEncoding.EncodeToString(binary)
	var documentID string

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a text document.
			{
				Config: testAccDocumentConfig("TLS cert", `    name    = "server.pem"
    content = "-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.doc", "category", "document"),
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.name", "server.pem"),
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.content", "-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"),
					resource.TestCheckNoResourceAttr("onepassword_item.doc", "document.content_base64"),
					resource.TestCheckResourceAttrSet("onepassword_item.doc", "document.id"),
					resource.TestCheckNoResourceAttr("onepassword_item.doc", "password"),
					resource.TestCheckResourceAttr("onepassword_item.doc", "note_value", "rotated yearly"),
					checkDocumentStored(items, t, "TLS cert", "server.pem", []byte("-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n")),
					checkReplaced(items, 0),
					captureAttr("onepassword_item.doc", "document.id", &documentID),
				),
			},
			// Changing only the title must not upload the file again.
			{
				Config: testAccDocumentConfig("TLS cert (prod)", `    name    = "server.pem"
    content = "-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.doc", "title", "TLS cert (prod)"),
					checkDocumentStored(items, t, "TLS cert (prod)", "server.pem", []byte("-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n")),
					checkReplaced(items, 0),
					sameAttr("onepassword_item.doc", "document.id", &documentID, true),
				),
			},
			// Changing the content replaces the file in place.
			{
				Config: testAccDocumentConfig("TLS cert (prod)", `    name    = "server.pem"
    content = "-----BEGIN CERTIFICATE-----\nBBBB\n-----END CERTIFICATE-----\n"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.content", "-----BEGIN CERTIFICATE-----\nBBBB\n-----END CERTIFICATE-----\n"),
					checkDocumentStored(items, t, "TLS cert (prod)", "server.pem", []byte("-----BEGIN CERTIFICATE-----\nBBBB\n-----END CERTIFICATE-----\n")),
					checkReplaced(items, 1),
					sameAttr("onepassword_item.doc", "document.id", &documentID, false),
					captureAttr("onepassword_item.doc", "document.id", &documentID),
				),
			},
			// Binary content and a new file name through content_base64.
			{
				Config: testAccDocumentConfig("TLS cert (prod)", fmt.Sprintf(`    name           = "keystore.p12"
    content_base64 = %q`, binaryBase64)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.name", "keystore.p12"),
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.content_base64", binaryBase64),
					resource.TestCheckNoResourceAttr("onepassword_item.doc", "document.content"),
					checkDocumentStored(items, t, "TLS cert (prod)", "keystore.p12", binary),
					checkReplaced(items, 2),
				),
			},
			// Import reads the file back, choosing content_base64 for binary data.
			{
				ResourceName:      "onepassword_item.doc",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccItemResourceDocument_DetectsDrift(t *testing.T) {
	items := newFakeSDKItems()
	useFakeSDK(t, items)

	config := testAccDocumentConfig("deploy key", `    name    = "id_ed25519"
    content = "original"`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				// Someone edits the file in the 1Password app.
				PreConfig:          func() { items.setDocumentContent(t, []byte("edited in the app")) },
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Applying puts the configured content back.
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDocumentStored(items, t, "deploy key", "id_ed25519", []byte("original")),
					checkReplaced(items, 1),
				),
			},
		},
	})
}

func TestAccItemResourceDocument_InvalidConfig(t *testing.T) {
	items := newFakeSDKItems()
	useFakeSDK(t, items)

	tests := map[string]struct {
		config string
		err    string
	}{
		"document on a login item": {
			config: testServiceAccountProviderConfig + fmt.Sprintf(`
resource "onepassword_item" "doc" {
  vault    = %q
  title    = "login"
  document = {
    name    = "a.txt"
    content = "a"
  }
}`, testDocumentVaultID),
			err: `'document' attribute can only be set when 'category' is "document"`,
		},
		"both content and content_base64": {
			config: testAccDocumentConfig("both", `    name           = "a.txt"
    content        = "a"
    content_base64 = "YQ=="`),
			err: `2 attributes specified when one \(and only one\) of`,
		},
		"neither content nor content_base64": {
			config: testAccDocumentConfig("neither", `    name = "a.txt"`),
			err:    `(?s)No attribute specified when one \(and only one\) of`,
		},
		"invalid base64": {
			config: testAccDocumentConfig("bad", `    name           = "a.bin"
    content_base64 = "not base64!"`),
			err: `Invalid base64 content`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(tc.err),
					},
				},
			})
		})
	}
}

func TestAccItemResourceDocument_ConnectNotSupported(t *testing.T) {
	expectedItem := generateBaseItem()
	expectedItem.Category = model.Document
	testServer := setupTestServer(&expectedItem, model.Vault{ID: expectedItem.VaultID, Name: "VaultName"}, t)
	defer testServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(testServer.URL) + fmt.Sprintf(`
resource "onepassword_item" "doc" {
  vault    = %q
  title    = "cert"
  category = "document"
  document = {
    name    = "server.pem"
    content = "pem"
  }
}`, expectedItem.VaultID),
				ExpectError: regexp.MustCompile(`not supported with 1Password Connect`),
			},
		},
	})
}

func TestAccItemResourceDocument_ContentKnownAfterApply(t *testing.T) {
	items := newFakeSDKItems()
	useFakeSDK(t, items)

	// The content comes from another resource, so it is unknown while planning.
	config := testServiceAccountProviderConfig + fmt.Sprintf(`
resource "terraform_data" "pem" {
  input = "generated"
}

resource "onepassword_item" "doc" {
  vault      = %q
  title      = "generated"
  category   = "document"
  note_value = "rotated yearly"

  document = {
    name    = "generated.pem"
    content = terraform_data.pem.output
  }
}
`, testDocumentVaultID)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.doc", "document.content", "generated"),
					checkDocumentStored(items, t, "generated", "generated.pem", []byte("generated")),
				),
			},
		},
	})
}
