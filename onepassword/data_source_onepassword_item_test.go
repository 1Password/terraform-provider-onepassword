package onepassword

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceOnePasswordItemRead(t *testing.T) {
	expectedItem := generateItem()
	meta := &testClient{
		GetItemFunc: func(uuid string, vaultUUID string) (*onepassword.Item, error) {
			return expectedItem, nil
		},
	}

	dataSourceData := generateDataSource(t, expectedItem)
	dataSourceData.Set("uuid", expectedItem.ID)

	err := dataSourceOnepasswordItemRead(context.Background(), dataSourceData, meta)
	if err != nil {
		t.Errorf("Unexpected error occured")
	}
	compareItemToSource(t, dataSourceData, expectedItem)
}

func TestDataSourceOnePasswordItemDocumentRead(t *testing.T) {
	expectedItem := generateItem()
	expectedItem.Category = "DOCUMENT"
	expectedItem.Files = []*onepassword.File{
		{
			Name: "test_file",
		},
	}
	expectedItem.Files[0].SetContent([]byte("test_content"))
	meta := &testClient{
		GetItemFunc: func(uuid string, vaultUUID string) (*onepassword.Item, error) {
			return expectedItem, nil
		},
		GetFileFunc: func(file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
			return []byte("test_content"), nil
		},
	}

	dataSourceData := generateDataSource(t, expectedItem)
	dataSourceData.Set("uuid", expectedItem.ID)

	err := dataSourceOnepasswordItemRead(context.Background(), dataSourceData, meta)
	if err != nil {
		t.Errorf("Unexpected error occured")
	}
	compareItemToSource(t, dataSourceData, expectedItem)
}

func TestDataSourceOnePasswordItemReadByTitle(t *testing.T) {
	expectedItem := generateItem()
	meta := &testClient{
		GetItemByTitleFunc: func(title string, vaultUUID string) (*onepassword.Item, error) {
			return expectedItem, nil
		},
	}

	dataSourceData := generateDataSource(t, expectedItem)
	dataSourceData.Set("title", expectedItem.Title)

	err := dataSourceOnepasswordItemRead(context.Background(), dataSourceData, meta)
	if err != nil {
		t.Errorf("Unexpected error occured")
	}
	compareItemToSource(t, dataSourceData, expectedItem)
}

func TestDataSourceOnePasswordItemReadWithSections(t *testing.T) {
	expectedItem := generateItem()
	meta := &testClient{
		GetItemFunc: func(uuid string, vaultUUID string) (*onepassword.Item, error) {
			return expectedItem, nil
		},
	}
	testSection := &onepassword.ItemSection{
		ID:    "1234",
		Label: "Test Section",
	}
	expectedItem.Sections = append(expectedItem.Sections, testSection)
	expectedItem.Fields = append(expectedItem.Fields, &onepassword.ItemField{
		ID:      "23456",
		Type:    "STRING",
		Label:   "Secret Information",
		Value:   "Password123",
		Section: testSection,
	})

	dataSourceData := generateDataSource(t, expectedItem)
	dataSourceData.Set("uuid", expectedItem.ID)

	err := dataSourceOnepasswordItemRead(context.Background(), dataSourceData, meta)
	if err != nil {
		t.Errorf("Unexpected error occured")
	}
	compareItemToSource(t, dataSourceData, expectedItem)
}

func compareItemToSource(t *testing.T, dataSourceData *schema.ResourceData, item *onepassword.Item) {
	if dataSourceData.Get("uuid") != item.ID {
		t.Errorf("Expected uuid to be %v got %v", item.ID, dataSourceData.Get("uuid"))
	}
	if dataSourceData.Get("vault") != item.Vault.ID {
		t.Errorf("Expected vault to be %v got %v", item.Vault.ID, dataSourceData.Get("vault"))
	}
	expectedCategory := strings.ToLower(fmt.Sprintf("%v", item.Category))
	if dataSourceData.Get("category") != expectedCategory {
		t.Errorf("Expected category to be %v got %v", expectedCategory, dataSourceData.Get("category"))
	}
	if dataSourceData.Get("title") != item.Title {
		t.Errorf("Expected title to be %v got %v", item.Title, dataSourceData.Get("title"))
	}
	if dataSourceData.Get("url") != item.URLs[0].URL {
		t.Errorf("Expected url to be %v got %v", item.URLs[0].URL, dataSourceData.Get("url"))
	}
	compareStringSlice(t, getTags(dataSourceData), item.Tags)

	for _, f := range item.Fields {
		path := f.Label
		if f.Section != nil {
			sectionIndex := 0
			fieldIndex := 0
			sections := dataSourceData.Get("section").([]interface{})

			for i, section := range sections {
				s := section.(map[string]interface{})
				if s["label"] == f.Section.Label ||
					(f.Section.ID != "" && s["id"] == f.Section.ID) {
					sectionIndex = i
					sectionFields := dataSourceData.Get(fmt.Sprintf("section.%d.field", i)).([]interface{})

					for j, field := range sectionFields {
						df := field.(map[string]interface{})
						if df["label"] == f.Label {
							fieldIndex = j
						}
					}
				}
			}

			if len(sections) > 0 {
				path = fmt.Sprintf("section.%d.field.%d.value", sectionIndex, fieldIndex)
			}
		}
		if dataSourceData.Get(path) != f.Value {
			t.Errorf("Expected field %v to be %v got %v", f.Label, f.Value, dataSourceData.Get(path))
		}
	}
	if files := dataSourceData.Get("file"); files != nil && len(item.Files) != len(files.([]interface{})) {
		got := len(files.([]interface{}))
		t.Errorf("Expected %v files got %v", len(item.Files), got)
	}
	for i, file := range item.Files {
		if dataSourceData.Get(fmt.Sprintf("file.%s", file.Name)) == nil {
			t.Errorf("Expected file %v to be present", file.Name)
		}
		want, err := file.Content()
		if err != nil {
			t.Errorf("Unexpected error occured")
		}
		if dataSourceData.Get(fmt.Sprintf("file.%d.content", i)).(string) != string(want) {
			t.Errorf("Expected file %v to have content %v, got %v", file.Name, string(want), dataSourceData.Get(fmt.Sprintf("file.%d.content", i)))
		}
	}
}

func generateDataSource(t *testing.T, item *onepassword.Item) *schema.ResourceData {
	dataSourceData := schema.TestResourceDataRaw(t, dataSourceOnepasswordItem().Schema, nil)
	dataSourceData.Set("vault", item.Vault.ID)
	dataSourceData.SetId(fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID))
	return dataSourceData
}

func generateItem() *onepassword.Item {
	item := onepassword.Item{}
	item.Fields = generateFields()
	item.ID = "79841a98-dd4a-4c34-8be5-32dca20a7328"
	item.Vault.ID = "df2e9643-45ad-4ff9-8b98-996f801afa75"
	item.Category = "USERNAME"
	item.Title = "test item"
	item.URLs = []onepassword.ItemURL{
		{
			Primary: true,
			URL:     "some_url.com",
		},
	}
	return &item
}

func generateFields() []*onepassword.ItemField {
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
			Value: "test_type",
		},
	}
	return fields
}

func compareStringSlice(t *testing.T, actual, expected []string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Expected slice to be length %d, but got %d", len(expected), len(actual))
		return
	}
	for i, val := range expected {
		if actual[i] != val {
			t.Errorf("Expected %s at index %d, but got %s", val, i, actual[i])
		}
	}
}
