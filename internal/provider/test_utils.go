package provider

import "github.com/1Password/connect-sdk-go/onepassword"

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
