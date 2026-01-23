// test/e2e/item_ephemeral_test.go
package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/cleanup"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/uuid"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccEphemeralItem_ReadAllItemTypes(t *testing.T) {
	t.Parallel()

	itemTypes := []struct {
		category model.ItemCategory
		name     string
		item     testItem
	}{
		{model.Login, "Login", testItems[model.Login]},
		{model.Password, "Password", testItems[model.Password]},
		{model.Database, "Database", testItems[model.Database]},
		{model.SecureNote, "SecureNote", testItems[model.SecureNote]},
		{model.Document, "Document", testItems[model.Document]},
		{model.SSHKey, "SSHKey", testItems[model.SSHKey]},
		{model.APICredential, "APICredential", testItems[model.APICredential]},
	}

	for _, itemType := range itemTypes {
		t.Run(itemType.name+"ByUUID", func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.EphemeralItemConfig(map[string]string{
								"vault": testVaultID,
								"uuid":  itemType.item.UUID,
							}),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							// Verify ephemeral resource is NOT in state
							resource.TestCheckFunc(func(s *terraform.State) error {
								ephemeralResource := s.RootModule().Resources["ephemeral.onepassword_item.test_item"]
								if ephemeralResource != nil {
									return fmt.Errorf("ephemeral resource should not exist in state, but found: %+v", ephemeralResource)
								}

								return nil
							}),
						),
					},
				},
			})
		})

		t.Run(itemType.name+"ByTitle", func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.EphemeralItemConfig(map[string]string{
								"vault": testVaultID,
								"title": itemType.item.Title,
							}),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							// Verify ephemeral resource is NOT in state
							resource.TestCheckFunc(func(s *terraform.State) error {
								ephemeralResource := s.RootModule().Resources["ephemeral.onepassword_item.test_item"]
								if ephemeralResource != nil {
									return fmt.Errorf("ephemeral resource should not exist in state, but found: %+v", ephemeralResource)
								}

								return nil
							}),
						),
					},
				},
			})
		})
	}
}

func TestAccEphemeralItem_UsePasswordInWriteOnlyField(t *testing.T) {
	t.Parallel()

	uniqueID := uuid.New().String()
	var copiedItemUUID string
	sourceItem := testItems[model.Password]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.EphemeralItemConfig(map[string]string{
						"vault": testVaultID,
						"uuid":  sourceItem.UUID,
					}),
					tfconfig.ItemResourceConfig(testVaultID, map[string]any{
						"title":               fmt.Sprintf("Ephemeral Test Copied %s", uniqueID),
						"category":            "login",
						"username":            "copieduser",
						"password_wo":         "ephemeral.onepassword_item.test_item.password",
						"password_wo_version": 1,
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", fmt.Sprintf("Ephemeral Test Copied %s", uniqueID)),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "copieduser"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &copiedItemUUID),
					cleanup.RegisterItem(t, &copiedItemUUID, testVaultID),
				),
			},
		},
	})
}

func TestAccEphemeralItem_UseNoteValueInWriteOnlyField(t *testing.T) {
	t.Parallel()

	uniqueID := uuid.New().String()
	var copiedItemUUID string
	sourceItem := testItems[model.SecureNote]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.EphemeralItemConfig(map[string]string{
						"vault": testVaultID,
						"uuid":  sourceItem.UUID,
					}),
					tfconfig.ItemResourceConfig(testVaultID, map[string]any{
						"title":                 fmt.Sprintf("Ephemeral Test Copied %s", uniqueID),
						"category":              "secure_note",
						"note_value_wo":         "ephemeral.onepassword_item.test_item.note_value",
						"note_value_wo_version": 1,
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", fmt.Sprintf("Ephemeral Test Copied %s", uniqueID)),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "secure_note"),
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &copiedItemUUID),
					cleanup.RegisterItem(t, &copiedItemUUID, testVaultID),
				),
			},
		},
	})
}

func TestAccEphemeralItem_NotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		config func() string
	}{
		{
			name: "ByUUID",
			config: func() string {
				return tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.EphemeralItemConfig(map[string]string{
						"vault": testVaultID,
						"uuid":  "nonexistent-uuid-12345",
					}),
				)
			},
		},
		{
			name: "ByTitle",
			config: func() string {
				return tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.EphemeralItemConfig(map[string]string{
						"vault": testVaultID,
						"title": "Nonexistent Item Title",
					}),
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config(),
						ExpectError: regexp.MustCompile("Unable to read item"),
					},
				},
			})
		})
	}
}
