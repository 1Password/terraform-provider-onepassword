package attributes

import (
	"fmt"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func BuildUpdatedItemAttrs(currentItem *model.Item, updatedAttrs map[string]any) *model.Item {
	return &model.Item{
		ID:       currentItem.ID,
		Title:    updatedAttrs["title"].(string),
		VaultID:  currentItem.VaultID,
		Category: currentItem.Category,
		Tags:     updatedAttrs["tags"].([]string),
		URLs: []model.ItemURL{
			{
				URL:     updatedAttrs["url"].(string),
				Primary: true,
			},
		},
		Sections: []model.ItemSection{
			{Label: "Updated Section", ID: "updated_section_id"},
			{Label: "Additional Section", ID: "additional_section_id"},
		},
		Fields: []model.ItemField{
			{ID: "username", Label: "username", Value: updatedAttrs["username"].(string), Type: "STRING", Purpose: "USERNAME"},
			{ID: "password", Label: "password", Value: updatedAttrs["password"].(string), Type: "CONCEALED", Purpose: "PASSWORD"},
			{ID: "notesPlain", Label: "notesPlain", Value: updatedAttrs["note_value"].(string), Type: "STRING", Purpose: "NOTES"},
			{Label: "New Field 3", Value: "new value 3", Type: "URL", SectionID: "updated_section_id"},
			{Label: "Extra Field", Value: "extra value", Type: "CONCEALED", SectionID: "additional_section_id"},
		},
	}
}

func BuildImportAttrs(attrs map[string]any) map[string]string {
	result := map[string]string{
		"url":        fmt.Sprintf("%v", attrs["url"]),
		"username":   fmt.Sprintf("%v", attrs["username"]),
		"password":   fmt.Sprintf("%v", attrs["password"]),
		"note_value": fmt.Sprintf("%v", attrs["note_value"]),
	}

	// Add tags
	if tags, ok := attrs["tags"].([]string); ok {
		for i, tag := range tags {
			result[fmt.Sprintf("tags.%d", i)] = tag
		}
	}

	// Add sections
	if sections, ok := attrs["section"].([]map[string]any); ok {
		for i, section := range sections {
			result[fmt.Sprintf("section.%d.label", i)] = section["label"].(string)

			// Extract fields
			fields, ok := section["field"].([]map[string]any)
			if ok {
				for j, field := range fields {
					prefix := fmt.Sprintf("section.%d.field.%d", i, j)
					result[fmt.Sprintf("%s.label", prefix)] = field["label"].(string)
					result[fmt.Sprintf("%s.value", prefix)] = field["value"].(string)
					result[fmt.Sprintf("%s.type", prefix)] = field["type"].(string)
				}
			}
		}
	}

	return result
}
