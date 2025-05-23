package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/1Password/connect-sdk-go/onepassword"
	"golang.org/x/crypto/ssh"
)

func generateBaseItem() onepassword.Item {
	item := onepassword.Item{}
	item.ID = "rix6gwgpuyog4gqplegvrp3dbm"
	item.Vault.ID = "gs2jpwmahszwq25a7jiw45e4je"
	item.Title = "test item"

	return item
}

func generateItemWithSections() *onepassword.Item {
	item := generateBaseItem()
	section := &onepassword.ItemSection{
		ID:    "1234",
		Label: "Test Section",
	}
	item.Sections = append(item.Sections, section)
	item.Fields = append(item.Fields, &onepassword.ItemField{
		ID:      "23456",
		Type:    "STRING",
		Label:   "Secret Information",
		Value:   "Password123",
		Section: section,
	})

	item.Category = onepassword.Login

	return &item
}

func generateDatabaseItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.Database
	item.Fields = generateDatabaseFields()

	return &item
}

func generateApiCredentialItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.Database
	item.Fields = generateApiCredentialFields()

	return &item
}

func generatePasswordItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.Password
	item.Fields = generatePasswordFields()

	return &item
}

func generateLoginItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.Login
	item.Fields = generateLoginFields()
	item.URLs = []onepassword.ItemURL{
		{
			Primary: true,
			URL:     "some_url.com",
		},
	}

	return &item
}

func generateSSHKeyItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.SSHKey
	item.Fields = generateSSHKeyFields()

	return &item
}

func generateSecureNoteItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.SecureNote
	item.Fields = []*onepassword.ItemField{
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Purpose: onepassword.FieldPurposeNotes,
			Value: `Lorem 
ipsum 
from 
notes
`,
		},
	}

	return &item
}

func generateDocumentItem() *onepassword.Item {
	item := generateBaseItem()
	item.Category = onepassword.Document
	item.Files = []*onepassword.File{
		{
			ID:          "ascii",
			Name:        "ascii",
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.Vault.ID, item.ID, "ascii"),
		},
		{
			ID:          "binary",
			Name:        "binary",
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.Vault.ID, item.ID, "binary"),
		},
	}
	item.Files[0].SetContent([]byte("ascii"))
	item.Files[1].SetContent([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	return &item
}

func generateLoginItemWithFiles() *onepassword.Item {
	item := generateItemWithSections()
	item.Category = onepassword.Login
	section := item.Sections[0]
	item.Files = []*onepassword.File{
		{
			ID:          "ascii",
			Name:        "ascii",
			Section:     section,
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.Vault.ID, item.ID, "ascii"),
		},
		{
			ID:          "binary",
			Name:        "binary",
			Section:     section,
			ContentPath: fmt.Sprintf("/v1/vaults/%s/items/%s/files/%s/content", item.Vault.ID, item.ID, "binary"),
		},
	}
	item.Files[0].SetContent([]byte("ascii"))
	item.Files[1].SetContent([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	return item
}

func generateDatabaseFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test_user",
		},
		{
			Label: "password",
			Value: "test_password",
		},
		{
			Label: "hostname",
			Value: "test_host",
		},
		{
			Label: "database",
			Value: "test_database",
		},
		{
			Label: "port",
			Value: "test_port",
		},
		{
			Label: "type",
			Value: "mysql",
		},
	}
	return fields
}

func generateApiCredentialFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test test_user",
		},
		{
			Label: "credential",
			Value: "test_credential",
		},
		{
			Label: "type",
			Value: "test_type",
		},
		{
			Label: "filename",
			Value: "test_filename",
		},
		{
			Label: "valid_from",
			Value: "test_valid_from",
		},
		{
			Label: "expires",
			Value: "expires_expires",
		},
		{
			Label: "hostname",
			Value: "test_hostname",
		},
	}
	return fields
}

func generatePasswordFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test_user",
		},
		{
			Label: "password",
			Value: "test_password",
		},
	}
	return fields
}

func generateLoginFields() []*onepassword.ItemField {
	fields := []*onepassword.ItemField{
		{
			Label: "username",
			Value: "test_user",
		},
		{
			Label: "password",
			Value: "test_password",
		},
	}
	return fields
}

func generateSSHKeyFields() []*onepassword.ItemField {
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

	fields := []*onepassword.ItemField{
		{
			Label: "private key",
			Value: string(pem.EncodeToMemory(privateKeyPem)),
		},
		{
			Label: "public key",
			Value: publicKey,
		},
	}
	return fields
}
