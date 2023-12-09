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

	for _, section := range item.Sections {
		t.Errorf("Missing Implementation for %s", section.Label)
		//keyedSection := dataSourceData.Get(fmt.Sprintf("keyed_sections.%s", section.Label)).(map[string]interface{})
		//dataSourceData.GetRawState()
		//if keyedSection == nil {
		//	t.Errorf("Expected keyed section %v to exist", section.Label)
		//}
		//if keyedSection["id"] != section.ID {
		//	t.Errorf("Expected keyed section %v to have id %v got %v", section.Label, section.ID, keyedSection["id"])
		//}
		//
		//for _, field := range item.Fields {
		//	if field.Section != nil && field.Section.ID == section.ID {
		//		keyedField := dataSourceData.Get(fmt.Sprintf("keyed_sections.%s.keyed_fields.%s", section.Label, field.Label)).(map[string]interface{})
		//
		//		if keyedField == nil {
		//			t.Errorf("Expected keyed field %v to exist", field.Label)
		//		}
		//		if keyedField["id"] != field.ID {
		//			t.Errorf("Expected keyed field %v to have id %v got %v", field.Label, field.ID, keyedField["id"])
		//		}
		//	}
		//}
	}

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
}

func generateDataSource(t *testing.T, item *onepassword.Item) *schema.ResourceData {
	dataSourceData := schema.TestResourceDataRaw(t, dataSourceOnepasswordItem().Schema, nil)
	dataSourceData.Set("vault", item.Vault.ID)
	dataSourceData.SetId(fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID))
	return dataSourceData
}

func generateItem() *onepassword.Item {
	fields, sections := generateFields()
	item := onepassword.Item{}
	item.Sections = sections
	item.Fields = fields
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

func generateFields() ([]*onepassword.ItemField, []*onepassword.ItemSection) {
	sections := []*onepassword.ItemSection{
		{
			ID:    "r9d6nt77oacycnjhp3tm3ihjka",
			Label: "Section 1",
			//Label: "", // seems to be the default section name
		},
		{
			ID:    "hbvbe4469kak5njjuthnhszcae",
			Label: "Section 2",
		},
	}

	// TODO: this is confusing, based on the JSON payload, fields live inside of sections there is only
	// payload.details.sections and inside of that there is a payload.details.sections.fields
	fields := []*onepassword.ItemField{
		{
			Label:   "username",
			Value:   "test_user",
			Section: sections[0],
		},
		{
			Label:   "password",
			Value:   "test_password",
			Section: sections[0],
		},
		{
			Label:   "hostname",
			Value:   "test_host",
			Section: sections[0],
		},
		{
			Label:   "database",
			Value:   "test_database",
			Section: sections[0],
		},
		{
			Label:   "port",
			Value:   "test_port",
			Section: sections[0],
			//Section: nil, TODO: why is Section a pointer when there HAVE to be a section for the field?
		},
		{
			Label:   "type",
			Value:   "test_type",
			Section: sections[1],
		},
	}

	return fields, sections
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
