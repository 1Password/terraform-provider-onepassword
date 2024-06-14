package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

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
