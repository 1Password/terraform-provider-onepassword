package integration

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/terraform/config"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/cleanup"
	testssh "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/ssh"
	uuidutil "github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/uuid"
	"github.com/1Password/terraform-provider-onepassword/v3/test/e2e/utils/vault"
)

func TestAccItemResourceSSHKeyCreation(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SSH Key Creation", uniqueID)

	createAttrs := map[string]any{
		"title":        title,
		"category":     "ssh_key",
		"ssh_key_type": "ed25519",
		"ssh_key_bits": 2048,
		"note_value":   "Test ssh key note",
	}

	var itemUUID string
	var privateKeyAfterCreate string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, createAttrs),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "ssh_key"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "ssh_key_type", "ed25519"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "key_type", "Ed25519"),
					resource.TestMatchResourceAttr("onepassword_item.test_item", "public_key", regexp.MustCompile(`^ssh-ed25519 `)),
					resource.TestMatchResourceAttr("onepassword_item.test_item", "fingerprint", regexp.MustCompile(`^SHA256:`)),
					resource.TestMatchResourceAttr("onepassword_item.test_item", "private_key", regexp.MustCompile(`BEGIN OPENSSH PRIVATE KEY`)),
					captureAttrValue("onepassword_item.test_item", "private_key", &privateKeyAfterCreate),
					validateKeyPair("onepassword_item.test_item"),
				),
			},
			// Read/import the item and verify it matches state.
			{
				ResourceName:            "onepassword_item.test_item",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("vaults/%s/items/%s", testVaultID, createAttrs["title"]),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"note_value"},
			},
			// Update the title; the key pair must not be rotated.
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, map[string]any{
						"title":        createAttrs["title"].(string) + " updated",
						"category":     "ssh_key",
						"ssh_key_type": "ed25519",
						"ssh_key_bits": 2048,
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("onepassword_item.test_item", "title", createAttrs["title"].(string)+" updated"),
					checkAttrValueUnchanged("onepassword_item.test_item", "private_key", &privateKeyAfterCreate),
					validateKeyPair("onepassword_item.test_item"),
				),
			},
		},
	})
}

func TestAccItemResourceSSHKeyCreationRSA(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()
	title := addUniqueIDToTitle("Test SSH Key Creation RSA", uniqueID)

	var itemUUID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, map[string]any{
						"title":        title,
						"category":     "ssh_key",
						"ssh_key_type": "rsa",
						"ssh_key_bits": 2048,
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "key_type", "RSA, 2048-bit"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "ssh_key_type", "rsa"),
					resource.TestCheckResourceAttr("onepassword_item.test_item", "ssh_key_bits", "2048"),
					resource.TestMatchResourceAttr("onepassword_item.test_item", "public_key", regexp.MustCompile(`^ssh-rsa `)),
					validateKeyPair("onepassword_item.test_item"),
				),
			},
		},
	})
}

func TestAccItemResourceMemorablePassword(t *testing.T) {
	testCases := []struct {
		name   string
		recipe map[string]any
		check  resource.TestCheckFunc
	}{
		{
			name: "Hyphenated",
			recipe: map[string]any{
				"type":       "memorable",
				"word_count": 3,
				"separator":  "hyphens",
			},
			check: resource.TestMatchResourceAttr("onepassword_item.test_item", "password", regexp.MustCompile(`^[A-Za-z]+(-[A-Za-z]+){2}$`)),
		},
		{
			name: "SpacesAndSyllables",
			recipe: map[string]any{
				"type":       "memorable",
				"word_count": 3,
				"separator":  "spaces",
				"word_list":  "syllables",
			},
			check: resource.TestMatchResourceAttr("onepassword_item.test_item", "password", regexp.MustCompile(`^[A-Za-z]+( [A-Za-z]+){2}$`)),
		},
	}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uniqueID := uuid.New().String()

			var itemUUID string

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, map[string]any{
								"title":           addUniqueIDToTitle("Test Memorable Password", uniqueID),
								"category":        "password",
								"password_recipe": []map[string]any{tc.recipe},
							}),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
							cleanup.RegisterItem(t, &itemUUID, testVaultID),
							tc.check,
						),
					},
				},
			})
		})
	}
}

func TestAccItemResourcePINPassword(t *testing.T) {
	testVaultID := vault.GetTestVaultID(t)
	uniqueID := uuid.New().String()

	var itemUUID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfconfig.CreateConfigBuilder()(
					tfconfig.ProviderConfig(),
					tfconfig.ItemResourceConfig(testVaultID, map[string]any{
						"title":    addUniqueIDToTitle("Test PIN Password", uniqueID),
						"category": "password",
						"password_recipe": []map[string]any{{
							"type":   "pin",
							"length": 6,
						}},
					}),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
					cleanup.RegisterItem(t, &itemUUID, testVaultID),
					resource.TestMatchResourceAttr("onepassword_item.test_item", "password", regexp.MustCompile(`^[0-9]{6}$`)),
				),
			},
		},
	})
}

func TestAccItemResourceNewCategories(t *testing.T) {
	testCases := []struct {
		name       string
		attrs      map[string]any
		checks     []resource.TestCheckFunc
		secondStep map[string]any
	}{
		{
			name: "APICredential",
			attrs: map[string]any{
				"title":      "Test API Credential",
				"category":   "api_credential",
				"username":   "test_user",
				"credential": "test-credential-token",
				"valid_from": "2026-01-01",
				"filename":   "credentials.json",
				"note_value": "Test API credential note",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "api_credential"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "test_user"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "credential", "test-credential-token"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "valid_from", "2026-01-01"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "filename", "credentials.json"),
			},
			secondStep: map[string]any{
				"title":      "Test API Credential updated",
				"category":   "api_credential",
				"username":   "updated_user",
				"credential": "updated-credential-token",
				"valid_from": "2026-02-01",
				"filename":   "credentials.json",
			},
		},
		{
			name: "Server",
			attrs: map[string]any{
				"title":                  "Test Server",
				"category":               "server",
				"username":               "test_user",
				"admin_console_url":      "https://console.example.com",
				"admin_console_username": "admin_user",
				"admin_console_password": "admin_password",
				"note_value":             "Test server note",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "server"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "username", "test_user"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "admin_console_url", "https://console.example.com"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "admin_console_username", "admin_user"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "admin_console_password", "admin_password"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "section.#", "0"),
			},
			secondStep: map[string]any{
				"title":                  "Test Server updated",
				"category":               "server",
				"username":               "test_user",
				"admin_console_url":      "https://console.example.com/v2",
				"admin_console_username": "admin_user2",
				"admin_console_password": "admin_password2",
			},
		},
		{
			name: "WirelessRouter",
			attrs: map[string]any{
				"title":             "Test Wireless Router",
				"category":          "wireless_router",
				"network_name":      "test-network",
				"server_address":    "192.168.1.1",
				"wireless_security": "WPA2",
				"wireless_password": "wireless_password",
				"note_value":        "Test router note",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "wireless_router"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "network_name", "test-network"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "server_address", "192.168.1.1"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "wireless_security", "WPA2"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "wireless_password", "wireless_password"),
			},
			secondStep: map[string]any{
				"title":             "Test Wireless Router updated",
				"category":          "wireless_router",
				"network_name":      "updated-network",
				"server_address":    "192.168.1.2",
				"wireless_security": "WPA3",
				"wireless_password": "updated_wireless_password",
			},
		},
		{
			name: "SoftwareLicense",
			attrs: map[string]any{
				"title":            "Test Software License",
				"category":         "software_license",
				"license_key":      "ABCD-1234-EFGH-5678",
				"version":          "1.2.3",
				"licensed_to":      "Test User",
				"registered_email": "test@example.com",
				"download_link":    "https://example.com/download",
				"note_value":       "Test license note",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("onepassword_item.test_item", "category", "software_license"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "license_key", "ABCD-1234-EFGH-5678"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "version", "1.2.3"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "licensed_to", "Test User"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "registered_email", "test@example.com"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "download_link", "https://example.com/download"),
				resource.TestCheckResourceAttr("onepassword_item.test_item", "section.#", "0"),
			},
			secondStep: map[string]any{
				"title":            "Test Software License updated",
				"category":         "software_license",
				"license_key":      "EFGH-5678-IJKL-9012",
				"version":          "2.0.0",
				"licensed_to":      "Updated User",
				"registered_email": "updated@example.com",
				"download_link":    "https://example.com/download-v2",
			},
		},
	}

	testVaultID := vault.GetTestVaultID(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uniqueID := uuid.New().String()

			var itemUUID string

			createAttrs := withUniqueTitle(tc.attrs, uniqueID)
			updatedAttrs := withUniqueTitle(tc.secondStep, uniqueID)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, createAttrs),
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							append([]resource.TestCheckFunc{
								uuidutil.CaptureItemUUID(t, "onepassword_item.test_item", &itemUUID),
								cleanup.RegisterItem(t, &itemUUID, testVaultID),
							}, tc.checks...)...,
						),
					},
					{
						Config: tfconfig.CreateConfigBuilder()(
							tfconfig.ProviderConfig(),
							tfconfig.ItemResourceConfig(testVaultID, updatedAttrs),
						),
					},
				},
			})
		})
	}
}

func withUniqueTitle(attrs map[string]any, uniqueID string) map[string]any {
	updated := make(map[string]any, len(attrs))
	for k, v := range attrs {
		updated[k] = v
	}
	if title, ok := attrs["title"].(string); ok {
		updated["title"] = addUniqueIDToTitle(title, uniqueID)
	}
	return updated
}

func captureAttrValue(resourceName, attrName string, valuePtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		*valuePtr = rs.Primary.Attributes[attrName]
		if *valuePtr == "" {
			return fmt.Errorf("attribute %s is empty", attrName)
		}
		return nil
	}
}

func checkAttrValueUnchanged(resourceName, attrName string, expectedPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		actual := rs.Primary.Attributes[attrName]
		if actual != *expectedPtr {
			return fmt.Errorf("attribute %s changed: expected the value captured in a previous step", attrName)
		}
		return nil
	}
}

func validateKeyPair(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		publicKey := rs.Primary.Attributes["public_key"]
		privateKey := rs.Primary.Attributes["private_key"]

		if err := testssh.ValidateSSHKeys(publicKey, privateKey); err != nil {
			return fmt.Errorf("generated key pair invalid: %w", err)
		}

		// The private key must not leak the PKCS#8 storage format.
		if strings.Contains(privateKey, "BEGIN PRIVATE KEY") {
			return fmt.Errorf("private_key is in PKCS#8 format, expected OpenSSH format")
		}

		return nil
	}
}
