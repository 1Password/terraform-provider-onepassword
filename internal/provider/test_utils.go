package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

func generateBaseItem() model.Item {
	item := model.Item{}
	item.ID = "rix6gwgpuyog4gqplegvrp3dbm"
	item.VaultID = "gs2jpwmahszwq25a7jiw45e4je"
	item.Title = "test item"

	return item
}

func generateItemWithSections() *model.Item {
	item := generateBaseItem()
	section := model.ItemSection{
		ID:    "1234",
		Label: "Test Section",
	}
	item.Sections = append(item.Sections, section)
	item.Fields = append(item.Fields, model.ItemField{
		ID:           "23456",
		Type:         "STRING",
		Label:        "Secret Information",
		Value:        "Password123",
		SectionID:    section.ID,
		SectionLabel: section.Label,
	})

	item.Category = model.Login

	return &item
}

func generateDatabaseItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Database
	item.Fields = generateDatabaseFields()

	return &item
}

func generateApiCredentialItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.APICredential
	item.Fields = generateApiCredentialFields()

	return &item
}

func generatePasswordItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Password
	item.Fields = generatePasswordFields()

	return &item
}

func generateLoginItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Login
	item.Fields = generateLoginFields()
	item.URLs = []model.ItemURL{
		{
			Primary: true,
			URL:     "some_url.com",
		},
	}

	return &item
}

func generateSSHKeyItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.SSHKey
	item.Fields = generateSSHKeyFields()

	return &item
}

func generateSecureNoteItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.SecureNote
	item.Fields = []model.ItemField{
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Purpose: model.FieldPurposeNotes,
			Value: `Lorem
ipsum
from
notes
`,
		},
	}

	return &item
}

func generateDocumentItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Document
	item.Files = []model.ItemFile{
		{
			ID:          "ascii",
			Name:        "ascii",
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.VaultID, item.ID, "ascii"),
		},
		{
			ID:          "binary",
			Name:        "binary",
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.VaultID, item.ID, "binary"),
		},
	}
	item.Files[0].SetContent([]byte("ascii"))
	item.Files[1].SetContent([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	return &item
}

func generateLoginItemWithFiles() *model.Item {
	item := generateItemWithSections()
	item.Category = model.Login
	section := item.Sections[0]
	item.Files = []model.ItemFile{
		{
			ID:           "ascii",
			Name:         "ascii",
			SectionID:    section.ID,
			SectionLabel: section.Label,
			ContentPath:  fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.VaultID, item.ID, "ascii"),
		},
		{
			ID:           "binary",
			Name:         "binary",
			SectionID:    section.ID,
			SectionLabel: section.Label,
			ContentPath:  fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.VaultID, item.ID, "binary"),
		},
	}
	item.Files[0].SetContent([]byte("ascii"))
	item.Files[1].SetContent([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	return item
}

func generateDatabaseFields() []model.ItemField {
	fields := []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Value: "test_user",
		},
		{
			ID:    "password",
			Label: "password",
			Value: "test_password",
		},
		{
			ID:    "hostname",
			Label: "hostname",
			Value: "test_host",
		},
		{
			ID:    "database",
			Label: "database",
			Value: "test_database",
		},
		{
			ID:    "port",
			Label: "port",
			Value: "test_port",
		},
		{
			ID:    "type",
			Label: "type",
			Value: "mysql",
		},
	}
	return fields
}

func generateApiCredentialFields() []model.ItemField {
	fields := []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Value: "test test_user",
		},
		{
			ID:    "credential",
			Label: "credential",
			Value: "test_credential",
		},
		{
			ID:    "type",
			Label: "type",
			Value: "test_type",
		},
		{
			ID:    "filename",
			Label: "filename",
			Value: "test_filename",
		},
		{
			ID:    "validFrom",
			Label: "valid_from",
			Value: "test_valid_from",
		},
		{
			ID:    "hostname",
			Label: "hostname",
			Value: "test_hostname",
		},
	}
	return fields
}

// generateApiCredentialResourceItem is a fixture shaped like what the
// resource creates for the api_credential category: typed fields with the
// template's field IDs, including a valid DATE timestamp.
func generateApiCredentialResourceItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.APICredential
	item.Fields = []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   "test_user",
		},
		{
			ID:    "credential",
			Label: "credential",
			Type:  model.FieldTypeConcealed,
			Value: "test_credential",
		},
		{
			ID:    "validFrom",
			Label: "valid from",
			Type:  model.FieldTypeDate,
			Value: "2026-01-01",
		},
		{
			ID:    "filename",
			Label: "filename",
			Type:  model.FieldTypeString,
			Value: "test_filename",
		},
	}

	return &item
}

func generatePasswordFields() []model.ItemField {
	fields := []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Value: "test_user",
		},
		{
			ID:    "password",
			Label: "password",
			Value: "test_password",
		},
	}
	return fields
}

func generateLoginFields() []model.ItemField {
	fields := []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Value: "test_user",
		},
		{
			ID:    "password",
			Label: "password",
			Value: "test_password",
		},
	}
	return fields
}

func generateSSHKeyFields() []model.ItemField {
	bitSize := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}
	privateKeyPem := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	publicRSAKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	publicKey := "ssh-rsa " + base64.StdEncoding.EncodeToString(publicRSAKey.Marshal())

	fields := []model.ItemField{
		{
			ID:    "private_key",
			Label: "private key",
			Value: string(pem.EncodeToMemory(privateKeyPem)),
		},
		{
			ID:    "public_key",
			Label: "public key",
			Value: publicKey,
		},
	}
	return fields
}

func generateServerItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Server
	item.Sections = []model.ItemSection{{ID: "admin_console", Label: "Admin Console"}}
	item.Fields = []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   "test_user",
		},
		{
			ID:      "password",
			Label:   "password",
			Purpose: model.FieldPurposePassword,
			Type:    model.FieldTypeConcealed,
			Value:   "test_password",
		},
		{
			ID:           "admin_console_url",
			Label:        "admin console URL",
			Type:         model.FieldTypeString,
			Value:        "https://console.example.com",
			SectionID:    "admin_console",
			SectionLabel: "Admin Console",
		},
		{
			ID:           "admin_console_username",
			Label:        "admin console username",
			Type:         model.FieldTypeString,
			Value:        "admin_user",
			SectionID:    "admin_console",
			SectionLabel: "Admin Console",
		},
		{
			ID:           "admin_console_password",
			Label:        "console password",
			Type:         model.FieldTypeConcealed,
			Value:        "admin_password",
			SectionID:    "admin_console",
			SectionLabel: "Admin Console",
		},
	}

	return &item
}

func generateRouterItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.Router
	item.Fields = []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Type:  model.FieldTypeString,
			Value: "test_user",
		},
		{
			ID:      "password",
			Label:   "base station password",
			Purpose: model.FieldPurposePassword,
			Type:    model.FieldTypeConcealed,
			Value:   "test_password",
		},
		{
			ID:    "server",
			Label: "server / IP address",
			Type:  model.FieldTypeString,
			Value: "192.168.1.1",
		},
		{
			ID:    "network_name",
			Label: "network name",
			Type:  model.FieldTypeString,
			Value: "test-network",
		},
		{
			ID:    "wireless_security",
			Label: "wireless security",
			Type:  model.FieldTypeMenu,
			Value: "WPA2",
		},
		{
			ID:    "wireless_password",
			Label: "wireless network password",
			Type:  model.FieldTypeConcealed,
			Value: "wireless_password",
		},
	}

	return &item
}

func generateSoftwareLicenseItem() *model.Item {
	item := generateBaseItem()
	item.Category = model.SoftwareLicense
	item.Sections = []model.ItemSection{
		{ID: "customer", Label: "Customer"},
		{ID: "publisher", Label: "Publisher"},
	}
	item.Fields = []model.ItemField{
		{
			ID:    "reg_code",
			Label: "license key",
			Type:  model.FieldTypeString,
			Value: "ABCD-1234-EFGH-5678",
		},
		{
			ID:    "product_version",
			Label: "version",
			Type:  model.FieldTypeString,
			Value: "1.2.3",
		},
		{
			ID:           "reg_name",
			Label:        "licensed to",
			Type:         model.FieldTypeString,
			Value:        "Test User",
			SectionID:    "customer",
			SectionLabel: "Customer",
		},
		{
			ID:           "reg_email",
			Label:        "registered email",
			Type:         model.FieldTypeEmail,
			Value:        "test@example.com",
			SectionID:    "customer",
			SectionLabel: "Customer",
		},
		{
			ID:           "download_link",
			Label:        "download page",
			Type:         model.FieldTypeURL,
			Value:        "https://example.com/download",
			SectionID:    "publisher",
			SectionLabel: "Publisher",
		},
	}

	return &item
}
