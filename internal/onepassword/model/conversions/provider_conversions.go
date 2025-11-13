package conversions

import (
	"strings"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	sdk "github.com/1password/onepassword-sdk-go"
)

// FromSDK creates a new Item from an SDK item
func FromSDKItem(item *sdk.Item) *model.Item {
	if item == nil {
		return nil
	}

	sectionMap := make(map[string]*model.ItemSection)

	providerItem := &model.Item{
		ID:       item.ID,
		Title:    item.Title,
		VaultID:  item.VaultID,
		Category: fromSDKCategory(string(item.Category)),
		Tags:     fromSDKTags(item.Tags),
		URLs:     fromSDKURLs(item.Websites),
		Sections: fromSDKSections(item, sectionMap),
		Fields:   fromSDKFields(item, sectionMap),
		Files:    fromSDKFiles(item),
	}

	// Add notes as a field if present
	if item.Notes != "" {
		providerItem.Fields = append(providerItem.Fields, &model.ItemField{
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   item.Notes,
		})
	}

	return providerItem
}

func fromSDKFieldType(sdkType sdk.ItemFieldType) model.ItemFieldType {
	switch sdkType {
	case "Text":
		return "STRING"
	case "SshKey":
		return "SSH_KEY"
	default:
		return model.ItemFieldType(strings.ToUpper(string(sdkType)))
	}
}

func fromSDKCategory(category string) model.ItemCategory {
	switch category {
	case "SecureNote":
		return model.ItemCategorySecureNote
	case "SshKey":
		return model.ItemCategorySSHKey
	default:
		return model.ItemCategory(category)
	}
}

func fromSDKURLs(websites []sdk.Website) []model.ItemURL {
	urls := make([]model.ItemURL, 0, len(websites))
	for idx, w := range websites {
		urls = append(urls, model.ItemURL{
			URL:     w.URL,
			Label:   w.Label,
			Primary: idx == 0,
		})
	}
	return urls
}

func fromSDKTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func fromSDKSections(item *sdk.Item, sectionMap map[string]*model.ItemSection) []*model.ItemSection {
	var sections []*model.ItemSection
	for _, s := range item.Sections {
		if s.ID != "" {
			section := &model.ItemSection{
				ID:    s.ID,
				Label: s.Title,
			}
			sections = append(sections, section)
			sectionMap[s.ID] = section
		}
	}
	return sections
}

func fromSDKFields(item *sdk.Item, sectionMap map[string]*model.ItemSection) []*model.ItemField {
	fields := make([]*model.ItemField, 0, len(item.Fields))

	for _, f := range item.Fields {
		field := &model.ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  fromSDKFieldType(f.FieldType),
			Value: f.Value,
		}

		// Set purpose based on field ID
		switch f.ID {
		case "username":
			field.Purpose = model.FieldPurposeUsername
		case "password":
			field.Purpose = model.FieldPurposePassword
		}

		// Associate field with section if applicable
		if f.SectionID != nil && *f.SectionID != "" {
			if section, exists := sectionMap[*f.SectionID]; exists {
				field.Section = section
			}
		}

		fields = append(fields, field)

		// Add SSH public key as separate field
		if f.Details != nil && f.FieldType == "SshKey" {
			if sshKey := f.Details.SSHKey(); sshKey != nil {
				fields = append(fields, &model.ItemField{
					ID:    "public_key",
					Label: "public key",
					Type:  model.FieldTypeString,
					Value: sshKey.PublicKey,
				})
			}
		}
	}

	return fields
}

func fromSDKFiles(item *sdk.Item) []*model.ItemFile {
	files := make([]*model.ItemFile, 0, len(item.Files)+1)

	for _, f := range item.Files {
		files = append(files, &model.ItemFile{
			ID:   f.Attributes.ID,
			Name: f.Attributes.Name,
			Size: int(f.Attributes.Size),
		})
	}

	// Append the document if it exists
	if item.Document != nil {
		files = append(files, &model.ItemFile{
			ID:   item.Document.ID,
			Name: item.Document.Name,
			Size: int(item.Document.Size),
		})
	}

	return files
}
