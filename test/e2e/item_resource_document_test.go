package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/cleanup"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/client"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/uuid"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/vault"
)

func documentResourceConfig(vaultID, title, documentAttrs string) func() string {
	return func() string {
		return fmt.Sprintf(`resource "onepassword_item" "test_document" {
  vault    = %q
  title    = %q
  category = "document"

  document = {
%s
  }
}`, vaultID, title, documentAttrs)
	}
}

// verifyDocumentIn1Password downloads the document file and compares it with the expected content.
func verifyDocumentIn1Password(vaultID string, itemUUID *string, expectedName string, expectedContent []byte) resource.TestCheckFunc {
	return func(*terraform.State) error {
		ctx := context.Background()
		c, err := client.CreateTestClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		item, err := c.GetItem(ctx, *itemUUID, vaultID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}
		if item.Document == nil {
			return fmt.Errorf("item %q has no document", *itemUUID)
		}
		if item.Document.Name != expectedName {
			return fmt.Errorf("document name mismatch: expected %q, got %q", expectedName, item.Document.Name)
		}

		content, err := c.GetFileContent(ctx, item.Document, item.ID, item.VaultID)
		if err != nil {
			return fmt.Errorf("failed to read document: %w", err)
		}
		if string(content) != string(expectedContent) {
			return fmt.Errorf("document content mismatch: expected %q, got %q", expectedContent, content)
		}
		return nil
	}
}

func TestAccItemResourceDocument(t *testing.T) {
	if os.Getenv("OP_CONNECT_TOKEN") != "" {
		t.Skip("Document items can't be created with 1Password Connect")
	}

	testVaultID := vault.GetTestVaultID(t)
	title := addUniqueIDToTitle("Test Document", uuid.New().String())
	binary := []byte{0x00, 0xff, 0x10, 0x80, 'p', '1', '2'}
	var itemUUID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					documentResourceConfig(testVaultID, title, `    name    = "server.pem"
    content = "first version"`),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_document", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_document", "category", "document"),
					resource.TestCheckResourceAttr("onepassword_item.test_document", "document.content", "first version"),
					resource.TestCheckResourceAttrSet("onepassword_item.test_document", "document.id"),
					verifyDocumentIn1Password(testVaultID, &itemUUID, "server.pem", []byte("first version")),
				),
			},
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					documentResourceConfig(testVaultID, title, fmt.Sprintf(`    name           = "keystore.p12"
    content_base64 = %q`, base64.StdEncoding.EncodeToString(binary))),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.VerifyItemUUIDUnchanged(t, "onepassword_item.test_document", &itemUUID),
					resource.TestCheckResourceAttr("onepassword_item.test_document", "document.name", "keystore.p12"),
					verifyDocumentIn1Password(testVaultID, &itemUUID, "keystore.p12", binary),
				),
			},
			{
				ResourceName:      "onepassword_item.test_document",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
