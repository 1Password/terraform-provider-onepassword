package model

import (
	"context"
	"fmt"

	sdk "github.com/1password/onepassword-sdk-go"
)

type CharacterSet string
type ItemCategory string
type ItemFieldPurpose string
type ItemFieldType string

const (
	CharacterSetDigits  CharacterSet = "DIGITS"
	CharacterSetSymbols CharacterSet = "SYMBOLS"

	Login      ItemCategory = "LOGIN"
	Password   ItemCategory = "PASSWORD"
	SecureNote ItemCategory = "SECURE_NOTE"
	Document   ItemCategory = "DOCUMENT"
	SSHKey     ItemCategory = "SSH_KEY"
	Database   ItemCategory = "DATABASE"

	FieldPurposeUsername ItemFieldPurpose = "USERNAME"
	FieldPurposePassword ItemFieldPurpose = "PASSWORD"
	FieldPurposeNotes    ItemFieldPurpose = "NOTES"

	FieldTypeConcealed ItemFieldType = "CONCEALED"
	FieldTypeDate      ItemFieldType = "DATE"
	FieldTypeEmail     ItemFieldType = "EMAIL"
	FieldTypeMenu      ItemFieldType = "MENU"
	FieldTypeMonthYear ItemFieldType = "MONTH_YEAR"
	FieldTypeOTP       ItemFieldType = "OTP"
	FieldTypeString    ItemFieldType = "STRING"
	FieldTypeURL       ItemFieldType = "URL"
)

type Item struct {
	ID       string
	Title    string
	VaultID  string
	Category ItemCategory
	Version  int
	Tags     []string
	URLs     []ItemURL
	Sections []ItemSection
	Fields   []ItemField
	Files    []ItemFile
}

type ItemSection struct {
	ID    string
	Label string
}

type ItemField struct {
	ID           string
	Label        string
	Type         ItemFieldType
	Value        string
	Purpose      ItemFieldPurpose
	SectionID    string
	SectionLabel string
	Recipe       *GeneratorRecipe
	Generate     bool
}

type GeneratorRecipe struct {
	Length        int
	CharacterSets []CharacterSet
}

type ItemURL struct {
	URL     string
	Label   string
	Primary bool
}

// FromSDKItemToModel creates a new Item from an SDK item
func (i *Item) FromSDKItemToModel(item *sdk.Item) error {
	if item == nil {
		return fmt.Errorf("cannot convert nil SDK item to model")
	}
	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.VaultID
	i.Category = ItemCategory(item.Category)
	i.Tags = item.Tags
	i.URLs = fromSDKURLs(item.Websites)

	// Convert sections/fields/files
	sectionMap := buildSectionMap(item)
	i.Sections = fromSDKSections(sectionMap)
	i.Files = fromSDKFiles(item, sectionMap)
	i.Fields = fromSDKFields(item, sectionMap)

	// Notes are stored top level in an item from the SDK
	if item.Notes != "" {
		i.Fields = append(i.Fields, ItemField{
			Type:    FieldTypeString,
			Purpose: FieldPurposeNotes,
			Value:   item.Notes,
		})
	}

	return nil
}

// FromModelItemToSDKCreateParams creates an SDK item create params from an Item
func (i *Item) FromModelItemToSDKCreateParams() sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  i.VaultID,
		Title:    i.Title,
		Category: sdk.ItemCategory(i.Category),
		Tags:     i.Tags,
		Sections: toSDKSections(i.Sections),
		Websites: toSDKWebsites(i.URLs),
	}

	params.Fields, params.Notes = toSDKFields(i.Fields)

	return params
}

func fromSDKURLs(websites []sdk.Website) []ItemURL {
	urls := make([]ItemURL, 0, len(websites))
	for idx, w := range websites {
		urls = append(urls, ItemURL{
			URL:     w.URL,
			Label:   w.Label,
			Primary: idx == 0,
		})
	}
	return urls
}

func fromSDKSections(sectionMap map[string]ItemSection) []ItemSection {
	sections := make([]ItemSection, 0, len(sectionMap))
	for _, section := range sectionMap {
		sections = append(sections, section)
	}
	return sections
}

func fromSDKFields(item *sdk.Item, sectionMap map[string]ItemSection) []ItemField {
	fields := make([]ItemField, 0, len(item.Fields))

	for _, f := range item.Fields {
		field := ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  ItemFieldType(f.FieldType),
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
			if section, exists := sectionMap[*f.SectionID]; exists {
				field.SectionID = section.ID
				field.SectionLabel = section.Label
			}
		}

		fields = append(fields, field)

		// The SDK SSH keys under a details section of the private key field
		// Add SSH public key as separate field
		if f.Details != nil && f.FieldType == sdk.ItemFieldTypeSSHKey {
			if sshKey := f.Details.SSHKey(); sshKey != nil {
				fields = append(fields, ItemField{
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

func fromSDKFiles(item *sdk.Item, sectionMap map[string]ItemSection) []ItemFile {
	// +1 to account for the document that may be appended at the end
	files := make([]ItemFile, 0, len(item.Files)+1)

	for _, f := range item.Files {
		file := ItemFile{
			ID:   f.Attributes.ID,
			Name: f.Attributes.Name,
			Size: int(f.Attributes.Size),
		}

		// Look up section by ID
		if f.SectionID != "" {
			if section, exists := sectionMap[f.SectionID]; exists {
				file.SectionID = section.ID
				file.SectionLabel = section.Label
			}
		}

		files = append(files, file)
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

func toSDKFields(fields []ItemField) ([]sdk.ItemField, *string) {
	var notes *string
	sdkFields := make([]sdk.ItemField, 0, len(fields))

	for _, field := range fields {
		if field.Purpose == FieldPurposeNotes {
			notes = &field.Value
			continue
		}
		sdkFields = append(sdkFields, toSDKField(field))
	}

	return sdkFields, notes
}

func toSDKField(f ItemField) sdk.ItemField {
	fieldID := f.ID

	if f.Generate && f.Recipe != nil {
		password, err := generatePassword(f.Recipe)
		if err == nil {
			f.Value = password
		} else {
			fmt.Printf("Error generating password: %v\n", err)
		}
	}

	field := sdk.ItemField{
		ID:        fieldID,
		Title:     f.Label,
		FieldType: sdk.ItemFieldType(f.Type),
		Value:     f.Value,
	}

	if f.SectionID != "" {
		field.SectionID = &f.SectionID
	}

	return field
}

func toSDKSections(sections []ItemSection) []sdk.ItemSection {
	sdkSections := make([]sdk.ItemSection, 0, len(sections))
	for _, section := range sections {
		sdkSections = append(sdkSections, sdk.ItemSection{
			ID:    section.ID,
			Title: section.Label,
		})
	}
	return sdkSections
}

func toSDKWebsites(urls []ItemURL) []sdk.Website {
	websites := make([]sdk.Website, 0, len(urls))
	for _, url := range urls {
		if url.URL != "" {
			websites = append(websites, sdk.Website{
				URL:   url.URL,
				Label: url.Label,
			})
		}
	}
	return websites
}

func generatePassword(recipe *GeneratorRecipe) (string, error) {
	includeDigits := false
	includeSymbols := false

	for _, characterSet := range recipe.CharacterSets {
		switch characterSet {
		case CharacterSetDigits:
			includeDigits = true
		case CharacterSetSymbols:
			includeSymbols = true
		}
	}

	passwordResponse, err := sdk.Secrets.GeneratePassword(
		context.Background(),
		sdk.NewPasswordRecipeTypeVariantRandom(&sdk.PasswordRecipeRandomInner{
			IncludeDigits:  includeDigits,
			IncludeSymbols: includeSymbols,
			Length:         uint32(recipe.Length),
		}),
	)
	if err != nil {
		return "", err
	}

	return passwordResponse.Password, nil
}

func buildSectionMap(item *sdk.Item) map[string]ItemSection {
	sectionMap := make(map[string]ItemSection, len(item.Sections))
	for _, s := range item.Sections {
		if s.ID != "" {
			sectionMap[s.ID] = ItemSection{
				ID:    s.ID,
				Label: s.Title,
			}
		}
	}
	return sectionMap
}
