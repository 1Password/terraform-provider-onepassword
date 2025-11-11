package types

import (
	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

func ToConnectItem(item *sdk.Item) *connect.Item {
	return &connect.Item{
		ID:       item.ID,
		Title:    item.Title,
		Vault:    connect.ItemVault{ID: item.VaultID},
		Category: connect.ItemCategory(item.Category),
		Tags:     item.Tags,
		Version:  int(item.Version),
		URLs:     websitesToConnect(item.Websites),
		Sections: sectionsToConnect(item.Sections),
		Fields:   fieldsToConnect(item.Fields, item.Notes, item.Sections),
	}
}

func websitesToConnect(websites []sdk.Website) []connect.ItemURL {
	urls := make([]connect.ItemURL, len(websites))
	for i, w := range websites {
		urls[i] = connect.ItemURL{
			URL:     w.URL,
			Label:   w.Label,
			Primary: i == 0,
		}
	}
	return urls
}

func sectionsToConnect(sections []sdk.ItemSection) []*connect.ItemSection {
	result := make([]*connect.ItemSection, len(sections))
	for i, s := range sections {
		result[i] = &connect.ItemSection{
			ID:    s.ID,
			Label: s.Title,
		}
	}
	return result
}

func fieldsToConnect(fields []sdk.ItemField, notes string, sections []sdk.ItemSection) []*connect.ItemField {
	// Build section lookup map
	sectionMap := make(map[string]*connect.ItemSection, len(sections))
	for i := range sections {
		sectionMap[sections[i].ID] = &connect.ItemSection{
			ID:    sections[i].ID,
			Label: sections[i].Title,
		}
	}

	result := make([]*connect.ItemField, 0, len(fields)+1)
	for _, f := range fields {
		field := &connect.ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  connect.ItemFieldType(f.FieldType),
			Value: f.Value,
		}

		switch f.ID {
		case "password":
			field.Purpose = connect.FieldPurposePassword
		case "username":
			field.Purpose = connect.FieldPurposeUsername
		}

		// Link to section
		if f.SectionID != nil {
			if section, ok := sectionMap[*f.SectionID]; ok {
				field.Section = section
			}
		}

		result = append(result, field)
	}

	// Add notes if present
	if notes != "" {
		result = append(result, &connect.ItemField{
			Type:    connect.ItemFieldType("STRING"),
			Purpose: connect.FieldPurposeNotes,
			Value:   notes,
		})
	}

	return result
}
