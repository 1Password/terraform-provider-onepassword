package model

import (
	sdk "github.com/1password/onepassword-sdk-go"
)

// FromSDK creates a new Item from an SDK item
func FromSDK(item *sdk.Item) *Item {
	if item == nil {
		return nil
	}

	sectionMap := make(map[string]*ItemSection)

	providerItem := &Item{
		ID:       item.ID,
		Title:    item.Title,
		VaultID:  item.VaultID,
		Category: toProviderCategory(string(item.Category)),
		Tags:     toSDKTags(item.Tags),
		URLs:     toProviderURLs(item.Websites),
		Sections: toProviderSections(item, sectionMap),
		Fields:   toProviderFields(item, sectionMap),
		Files:    toProviderFiles(item),
	}

	// Add notes as a field if present
	if item.Notes != "" {
		providerItem.Fields = append(providerItem.Fields, &ItemField{
			Type:    FieldTypeString,
			Purpose: FieldPurposeNotes,
			Value:   item.Notes,
		})
	}

	return providerItem
}

func (i *Item) ToSDK(vaultID string) sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  vaultID,
		Title:    i.Title,
		Category: toSDKCategory(i.Category),
		Tags:     toSDKTags(i.Tags),
	}

	// Convert fields
	for _, field := range i.Fields {
		if field.Purpose == FieldPurposeNotes {
			params.Notes = &field.Value
			continue
		}
		params.Fields = append(params.Fields, toSDKField(field))
	}

	// Convert sections
	for _, section := range i.Sections {
		params.Sections = append(params.Sections, sdk.ItemSection{
			ID:    section.ID,
			Title: section.Label,
		})
	}

	// Convert URLs
	for _, url := range i.URLs {
		if url.URL != "" {
			params.Websites = append(params.Websites, sdk.Website{
				URL:   url.URL,
				Label: url.Label,
			})
		}
	}

	return params
}
