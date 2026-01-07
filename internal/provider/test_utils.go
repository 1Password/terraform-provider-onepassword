package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
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
			ID:    "valid_from",
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
