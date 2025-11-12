package model

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/1password/onepassword-sdk-go"
	"github.com/hashicorp/go-uuid"
)

func populateSDKFields(f *ItemField) sdk.ItemField {
	fieldID := f.ID

	// connect generate uuid, but sdk does not
	if fieldID == "" {
		fieldID, _ = uuid.GenerateUUID()
	}

	if f.Generate && f.Recipe != nil {
		password, err := populatePasswordRecipe(f.Recipe)
		if err == nil {
			f.Value = password
		} else {
			fmt.Printf("Error generating password: %v\n", err)
		}
	}

	field := sdk.ItemField{
		ID:        fieldID,
		Title:     f.Label,
		FieldType: populateSDKFieldType(f.Type),
		Value:     f.Value,
	}

	if f.Section != nil {
		sectionID := f.Section.ID
		field.SectionID = &sectionID
	}

	return field
}

func populatePasswordRecipe(recipe *GeneratorRecipe) (string, error) {
	includeDigits := false
	includeSymbols := false

	for _, characterSet := range recipe.CharacterSets {
		if characterSet == "DIGITS" {
			includeDigits = true
		}
		if characterSet == "SYMBOLS" {
			includeSymbols = true
		}
	}

	passwordResponse, err := sdk.Secrets.GeneratePassword(context.Background(), sdk.NewPasswordRecipeTypeVariantRandom(&sdk.PasswordRecipeRandomInner{
		IncludeDigits:  includeDigits,
		IncludeSymbols: includeSymbols,
		Length:         uint32(recipe.Length),
	}))

	if err != nil {
		return "", err
	}

	return passwordResponse.Password, nil
}

func populateSDKFieldType(modelType ItemFieldType) sdk.ItemFieldType {
	switch modelType {
	case "string", "STRING":
		return "Text"
	case "concealed", "CONCEALED":
		return "Concealed"
	case "email", "EMAIL":
		return "Email"
	case "url", "URL":
		return "Url"
	case "date", "DATE":
		return "Date"
	case "SSH_KEY":
		return "SshKey"
	default:
		return sdk.ItemFieldType(modelType)
	}
}

func populateSDKCategoryType(modelType ItemCategory) sdk.ItemCategory {
	switch modelType {
	case ItemCategorySecureNote:
		return "SecureNote"
	case ItemCategorySSHKey:
		return "SshKey"
	default:
		return sdk.ItemCategory(modelType)
	}
}

func populateItemTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func populateProviderItemURLs(websites []sdk.Website) []ItemURL {
	var urls []ItemURL
	for idx, w := range websites {

		urls = append(urls, ItemURL{
			URL:     w.URL,
			Label:   w.Label,
			Primary: idx == 0,
		})
	}
	return urls
}

func populateProviderItemSections(item *sdk.Item, sectionMap map[string]*ItemSection) []*ItemSection {
	var sections []*ItemSection
	for _, s := range item.Sections {
		if s.ID != "" {
			section := &ItemSection{
				ID:    s.ID,
				Label: s.Title,
			}
			sections = append(sections, section)
			sectionMap[s.ID] = section
		}
	}
	return sections
}

func populateProviderItemFields(item *sdk.Item, sectionMap map[string]*ItemSection) []*ItemField {
	var fields []*ItemField
	for _, f := range item.Fields {
		field := &ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  populateProviderFieldType(f.FieldType),
			Value: f.Value,
		}

		// Set purpose based on field ID
		switch f.ID {
		case "username":
			field.Purpose = FieldPurposeUsername
		case "password":
			field.Purpose = FieldPurposePassword
		}

		// Associate field with section if applicable
		if f.SectionID != nil && *f.SectionID != "" {
			field.Section = sectionMap[*f.SectionID]
		}

		fields = append(fields, field)

		if f.Details != nil && f.FieldType == "SshKey" {
			if sshKey := f.Details.SSHKey(); sshKey != nil {
				fields = append(fields, &ItemField{
					ID:    "public_key",
					Label: "public key",
					Type:  FieldTypeString,
					Value: sshKey.PublicKey,
				})
			}
		}

	}

	return fields
}

func populateProviderFiles(item *sdk.Item) []ItemFile {
	var files []ItemFile

	for _, f := range item.Files {
		files = append(files, ItemFile{
			ID:   f.Attributes.ID,
			Name: f.Attributes.Name,
			Size: int(f.Attributes.Size),
		})
	}

	// Append the document if it exists
	if item.Document != nil {
		files = append(files, ItemFile{
			ID:   item.Document.ID,
			Name: item.Document.Name,
			Size: int(item.Document.Size),
		})
	}

	return files
}

func populateProviderFieldType(sdkType sdk.ItemFieldType) ItemFieldType {
	switch sdkType {
	case "Text":
		return "STRING"
	case "SshKey":
		return "SSH_KEY"
	default:
		return ItemFieldType(strings.ToUpper(string(sdkType)))
	}
}

func populateProviderCategory(category string) ItemCategory {
	switch category {
	case "SecureNote":
		return ItemCategorySecureNote
	case "SshKey":
		return ItemCategorySSHKey
	default:
		return ItemCategory(category)
	}
}
