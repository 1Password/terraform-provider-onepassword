package model

import (
	"context"
	"fmt"

	sdk "github.com/1password/onepassword-sdk-go"
	"github.com/hashicorp/go-uuid"
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

// FromSDK creates a new Item from an SDK item
func (i *Item) FromSDKItem(item *sdk.Item) {
	if item == nil {
		return
	}

	sectionMap := make(map[string]*ItemSection)

	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.VaultID
	i.Category = item.Category
	i.Tags = fromSDKTags(item.Tags)
	i.URLs = fromSDKURLs(item.Websites)
	i.Sections = fromSDKSections(item, sectionMap)
	i.Fields = fromSDKFields(item, sectionMap)
	i.Files = fromSDKFiles(item)

	if item.Notes != "" {
		i.Fields = append(i.Fields, &ItemField{
			Type:    sdk.ItemFieldTypeText,
			Purpose: FieldPurposeNotes,
			Value:   item.Notes,
		})
	}
}

func (i *Item) ToSDKItem(vaultID string) sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  vaultID,
		Title:    i.Title,
		Category: i.Category,
		Tags:     i.Tags,
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

func fromSDKTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	return tags
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

func toSDKField(f *ItemField) sdk.ItemField {
	fieldID := f.ID

	// connect generate uuid, but sdk does not
	if fieldID == "" {
		fieldID, _ = uuid.GenerateUUID()
	}

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

func toSDKTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	return tags
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
