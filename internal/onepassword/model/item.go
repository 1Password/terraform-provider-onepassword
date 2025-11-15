package model

import (
	"context"
	"fmt"

	sdk "github.com/1password/onepassword-sdk-go"
)

type ItemFieldPurpose string

const (
	FieldPurposeUsername ItemFieldPurpose = "USERNAME"
	FieldPurposePassword ItemFieldPurpose = "PASSWORD"
	FieldPurposeNotes    ItemFieldPurpose = "NOTES"
)

type Item struct {
	ID       string
	Title    string
	VaultID  string
	Category sdk.ItemCategory
	Version  int
	Tags     []string
	URLs     []ItemURL
	Sections []*ItemSection
	Fields   []*ItemField
	Files    []*ItemFile
}

type ItemSection struct {
	ID    string
	Label string
}

type ItemField struct {
	ID       string
	Label    string
	Type     sdk.ItemFieldType
	Value    string
	Purpose  ItemFieldPurpose
	Section  *ItemSection
	Recipe   *GeneratorRecipe
	Generate bool
}

type GeneratorRecipe struct {
	Length        int
	CharacterSets []string
}

type ItemURL struct {
	URL     string
	Label   string
	Primary bool
}

// FromSDKItemToModel creates a new Item from an SDK item
func (i *Item) FromSDKItemToModel(item *sdk.Item) {
	if item == nil {
		return
	}

	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.VaultID
	i.Category = item.Category
	i.Tags = item.Tags
	i.URLs = fromSDKURLs(item.Websites)
	i.Files = fromSDKFiles(item)

	// Convert sections and fields
	sectionMap := make(map[string]*ItemSection)
	i.Sections = fromSDKSections(item, sectionMap)
	i.Fields = fromSDKFields(item, sectionMap)

	// Notes are stored top level in an item from the SDK
	if item.Notes != "" {
		i.Fields = append(i.Fields, &ItemField{
			Type:    sdk.ItemFieldTypeText,
			Purpose: FieldPurposeNotes,
			Value:   item.Notes,
		})
	}
}

// FromModelItemToSDKCreateParams creates an SDK item create params from an Item
func (i *Item) FromModelItemToSDKCreateParams() sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  i.VaultID,
		Title:    i.Title,
		Category: i.Category,
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

func fromSDKSections(item *sdk.Item, sectionMap map[string]*ItemSection) []*ItemSection {
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

func fromSDKFields(item *sdk.Item, sectionMap map[string]*ItemSection) []*ItemField {
	fields := make([]*ItemField, 0, len(item.Fields))

	for _, f := range item.Fields {
		field := &ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  f.FieldType,
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
				field.Section = section
			}
		}

		fields = append(fields, field)

		// The SDK SSH keys under a details section of the private key field
		// Add SSH public key as separate field
		if f.Details != nil && f.FieldType == sdk.ItemFieldTypeSSHKey {
			if sshKey := f.Details.SSHKey(); sshKey != nil {
				fields = append(fields, &ItemField{
					ID:    "public_key",
					Label: "public key",
					Type:  sdk.ItemFieldTypeText,
					Value: sshKey.PublicKey,
				})
			}
		}
	}

	return fields
}

func fromSDKFiles(item *sdk.Item) []*ItemFile {
	files := make([]*ItemFile, 0, len(item.Files)+1)

	for _, f := range item.Files {
		files = append(files, &ItemFile{
			ID:   f.Attributes.ID,
			Name: f.Attributes.Name,
			Size: int(f.Attributes.Size),
		})
	}

	// Append the document if it exists
	if item.Document != nil {
		files = append(files, &ItemFile{
			ID:   item.Document.ID,
			Name: item.Document.Name,
			Size: int(item.Document.Size),
		})
	}

	return files
}

func toSDKFields(fields []*ItemField) ([]sdk.ItemField, *string) {
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

func toSDKField(f *ItemField) sdk.ItemField {
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
		FieldType: f.Type,
		Value:     f.Value,
	}

	if f.Section != nil {
		sectionID := f.Section.ID
		field.SectionID = &sectionID
	}

	return field
}

func toSDKSections(sections []*ItemSection) []sdk.ItemSection {
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
		case "DIGITS":
			includeDigits = true
		case "SYMBOLS":
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
