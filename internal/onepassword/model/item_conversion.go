package model

import (
	sdk "github.com/1password/onepassword-sdk-go"
)

func (i *Item) ConvertSDKItemToProviderItem(item *sdk.Item) {
	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.VaultID
	i.Category = populateProviderCategory(string(item.Category))
	i.Tags = populateItemTags(item.Tags)
	i.URLs = populateProviderItemURLs(item.Websites)

	// Sections and Fields
	sectionMap := make(map[string]*ItemSection)
	i.Sections = populateProviderItemSections(item, sectionMap)
	i.Fields = populateProviderItemFields(item, sectionMap)

	// Add notes as a field as it appears top level in the SDK item
	if item.Notes != "" {
		i.Fields = append(i.Fields, &ItemField{
			Type:    FieldTypeString,
			Purpose: FieldPurposeNotes,
			Value:   item.Notes,
		})
	}

	i.Files = populateProviderFiles(item)

}

func (i *Item) ConvertItemToSDKItem(vaultID string) sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  vaultID,
		Title:    i.Title,
		Category: populateSDKCategoryType(i.Category),
		Tags:     populateItemTags(i.Tags),
	}

	// Convert fields
	for _, f := range i.Fields {
		if f.Purpose == FieldPurposeNotes {
			params.Notes = &f.Value
			continue
		}

		field := populateSDKFields(f)
		params.Fields = append(params.Fields, field)
	}

	// Convert sections
	for _, s := range i.Sections {
		params.Sections = append(params.Sections, sdk.ItemSection{
			ID:    s.ID,
			Title: s.Label,
		})
	}

	// Convert URLs
	for _, u := range i.URLs {
		if u.URL != "" {
			params.Websites = append(params.Websites, sdk.Website{
				URL:   u.URL,
				Label: u.Label,
			})
		}
	}

	return params
}
