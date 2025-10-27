package integration

import (
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/config"
	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccItemResource(t *testing.T) {
	serviceAccountToken, err := config.GetServiceAccountToken()
	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	testCases := []struct {
		name               string
		itemResourceConfig tfconfig.ItemResource
		expectedAttributes map[string]string
		validateFunc       func(*terraform.State) error
	}{
		{
			name: "CreateLogin",
			itemResourceConfig: tfconfig.ItemResource{
				Auth: tfconfig.AuthConfig{
					ServiceAccountToken: serviceAccountToken,
				},
				Params: map[string]string{
					"vault":      "t7dnwbjh6nlyw475wl3m442sdi",
					"title":      "Test Login Item 2",
					"category":   "login",
					"username":   "testuser@example.com",
					"url":        "https://example.com",
					"password":   "testpassword",
					"note_value": "Test note",
				},
				Tags: []string{"tag1", "tag2"},
			},
			expectedAttributes: map[string]string{
				"title":      "Test Login Item 2",
				"category":   "login",
				"username":   "testuser@example.com",
				"url":        "https://example.com",
				"password":   "testpassword",
				"note_value": "Test note",
				"tags.0":     "tag1",
				"tags.1":     "tag2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resourceBuilder := tfconfig.CreateItemResourceConfigBuilder()

			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "vault", "t7dnwbjh6nlyw475wl3m442sdi"),
			}

			for attr, expectedValue := range tc.expectedAttributes {
				checks = append(checks, resource.TestCheckResourceAttr("onepassword_item.test_item", attr, expectedValue))
			}

			//	checks = append(checks, resource.TestCheckResourceAttrSet("onepassword_item.test_item", "uuid"))

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: resourceBuilder(
							tfconfig.ProviderAuthWithServiceAccount(tc.itemResourceConfig.Auth),
							tfconfig.ItemResourceConfig(tc.itemResourceConfig.Params, tc.itemResourceConfig.Tags),
						),
						Check: resource.ComposeAggregateTestCheckFunc(checks...),
					},
				},
			})
		})
	}
}
