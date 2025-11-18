package model

import (
	"context"
	"fmt"

	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"
)

type Item struct {
	ID       string               `json:"id"`
	Title    string               `json:"title"`
	VaultID  string               `json:"vaultId"`
	Category connect.ItemCategory `json:"category,omitempty"`
	Version  int                  `json:"version,omitempty"`
	Tags     []string             `json:"tags,omitempty"`
	URLs     []ItemURL            `json:"urls,omitempty"`
	Sections []ItemSection        `json:"sections,omitempty"`
	Fields   []ItemField          `json:"fields,omitempty"`
	Files    []ItemFile           `json:"files,omitempty"`
}

type ItemSection struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

type ItemField struct {
	ID       string                   `json:"id"`
	Label    string                   `json:"label,omitempty"`
	Type     connect.ItemFieldType    `json:"type"`
	Value    string                   `json:"value,omitempty"`
	Purpose  connect.ItemFieldPurpose `json:"purpose,omitempty"`
	Section  ItemSection              `json:"section,omitempty"`
	Recipe   *GeneratorRecipe         `json:"recipe,omitempty"`
	Generate bool                     `json:"generate,omitempty"`
}

type GeneratorRecipe struct {
	Length        int      `json:"length,omitempty"`
	CharacterSets []string `json:"characterSets,omitempty"`
}

type ItemURL struct {
	URL     string `json:"href"`
	Label   string `json:"label,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// FromSDKItemToModel creates a new Item from an SDK item
func (i *Item) FromSDKItemToModel(item *sdk.Item) {
	if item == nil {
		return
	}

	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.VaultID
	i.Category = connect.ItemCategory(item.Category)
	i.Tags = item.Tags
	i.URLs = fromSDKURLs(item.Websites)
	i.Files = fromSDKFiles(item)

	// Convert sections and fields
	sectionMap := make(map[string]ItemSection)
	i.Sections = fromSDKSections(item, sectionMap)
	i.Fields = fromSDKFields(item, sectionMap)

	// Notes are stored top level in an item from the SDK
	if item.Notes != "" {
		i.Fields = append(i.Fields, ItemField{
			Type:    connect.FieldTypeString,
			Purpose: connect.FieldPurposeNotes,
			Value:   item.Notes,
		})
	}
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

// FromConnectItemToModel creates a new Item from a Connect SDK item
func (i *Item) FromConnectItemToModel(item *connect.Item) {
	if item == nil {
		return
	}

	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.Vault.ID
	i.Category = item.Category
	i.Version = item.Version
	i.Tags = item.Tags
	i.URLs = fromConnectURLs(item.URLs)
	i.Files = fromConnectFiles(item.Files)

	// Convert sections and fields
	sectionMap := make(map[string]ItemSection)
	i.Sections = fromConnectSections(item.Sections, sectionMap)
	i.Fields = fromConnectFields(item.Fields, sectionMap)
}

// FromModelItemToConnect creates a Connect SDK item from a model Item
func (i *Item) FromModelItemToConnect() *connect.Item {
	return &connect.Item{
		ID:       i.ID,
		Title:    i.Title,
		Vault:    connect.ItemVault{ID: i.VaultID},
		Category: i.Category,
		Version:  i.Version,
		Tags:     i.Tags,
		URLs:     toConnectURLs(i.URLs),
		Sections: toConnectSections(i.Sections),
		Fields:   toConnectFields(i.Fields),
		Files:    toConnectFiles(i.Files),
	}
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

func fromSDKSections(item *sdk.Item, sectionMap map[string]ItemSection) []ItemSection {
	var sections []ItemSection
	for _, s := range item.Sections {
		if s.ID != "" {
			section := ItemSection{
				ID:    s.ID,
				Label: s.Title,
			}
			sections = append(sections, section)
			sectionMap[s.ID] = section
		}
	}
	return sections
}

func fromSDKFields(item *sdk.Item, sectionMap map[string]ItemSection) []ItemField {
	fields := make([]ItemField, 0, len(item.Fields))

	for _, f := range item.Fields {
		field := ItemField{
			ID:    f.ID,
			Label: f.Title,
			Type:  connect.ItemFieldType(f.FieldType),
			Value: f.Value,
		}

		// Set purpose based on field ID
		switch f.ID {
		case "username":
			field.Purpose = connect.FieldPurposeUsername
		case "password":
			field.Purpose = connect.FieldPurposePassword
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
				fields = append(fields, ItemField{
					ID:    "public_key",
					Label: "public key",
					Type:  connect.FieldTypeString,
					Value: sshKey.PublicKey,
				})
			}
		}
	}

	return fields
}

func fromSDKFiles(item *sdk.Item) []ItemFile {
	files := make([]ItemFile, 0, len(item.Files)+1)

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

func toSDKFields(fields []ItemField) ([]sdk.ItemField, *string) {
	var notes *string
	sdkFields := make([]sdk.ItemField, 0, len(fields))

	for _, field := range fields {
		if field.Purpose == connect.FieldPurposeNotes {
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

	if f.Section.ID != "" {
		field.SectionID = &f.Section.ID
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

func fromConnectURLs(urls []connect.ItemURL) []ItemURL {
	modelURLs := make([]ItemURL, 0, len(urls))
	for _, u := range urls {
		modelURLs = append(modelURLs, ItemURL{
			URL:     u.URL,
			Label:   u.Label,
			Primary: u.Primary,
		})
	}
	return modelURLs
}

func fromConnectSections(sections []*connect.ItemSection, sectionMap map[string]ItemSection) []ItemSection {
	modelSections := make([]ItemSection, 0, len(sections))
	for _, s := range sections {
		if s != nil {
			section := ItemSection{
				ID:    s.ID,
				Label: s.Label,
			}
			modelSections = append(modelSections, section)
			sectionMap[s.ID] = section
		}
	}
	return modelSections
}

func fromConnectFields(fields []*connect.ItemField, sectionMap map[string]ItemSection) []ItemField {
	modelFields := make([]ItemField, 0, len(fields))
	for _, f := range fields {
		if f == nil {
			continue
		}

		field := ItemField{
			ID:      f.ID,
			Label:   f.Label,
			Type:    f.Type,
			Value:   f.Value,
			Purpose: f.Purpose,
		}

		// Associate field with section if applicable
		if f.Section != nil && f.Section.ID != "" {
			if section, exists := sectionMap[f.Section.ID]; exists {
				field.Section = section
			}
		}

		// Handle password recipe if present
		if f.Recipe != nil {
			field.Recipe = &GeneratorRecipe{
				Length:        f.Recipe.Length,
				CharacterSets: f.Recipe.CharacterSets,
			}
		}

		modelFields = append(modelFields, field)
	}
	return modelFields
}

func fromConnectFiles(files []*connect.File) []ItemFile {
	result := make([]ItemFile, 0, len(files))
	for _, f := range files {
		if f == nil {
			continue
		}

		itemFile := ItemFile{
			ID:          f.ID,
			Name:        f.Name,
			Size:        f.Size,
			ContentPath: f.ContentPath,
		}

		// Only set Section if it exists
		if f.Section != nil {
			itemFile.Section = &ItemSection{
				ID:    f.Section.ID,
				Label: f.Section.Label,
			}
		}

		result = append(result, itemFile)
	}
	return result
}

func toConnectURLs(urls []ItemURL) []connect.ItemURL {
	connectURLs := make([]connect.ItemURL, 0, len(urls))
	for _, u := range urls {
		connectURLs = append(connectURLs, connect.ItemURL{
			URL:     u.URL,
			Label:   u.Label,
			Primary: u.Primary,
		})
	}
	return connectURLs
}

func toConnectSections(sections []ItemSection) []*connect.ItemSection {
	connectSections := make([]*connect.ItemSection, 0, len(sections))
	for _, s := range sections {
		connectSections = append(connectSections, &connect.ItemSection{
			ID:    s.ID,
			Label: s.Label,
		})
	}
	return connectSections
}

func toConnectFields(fields []ItemField) []*connect.ItemField {
	connectFields := make([]*connect.ItemField, 0, len(fields))
	for _, f := range fields {
		if f.Generate && f.Recipe != nil {
			password, err := generatePassword(f.Recipe)
			if err == nil {
				f.Value = password
			} else {
				fmt.Printf("Error generating password: %v\n", err)
			}
		}

		field := &connect.ItemField{
			ID:      f.ID,
			Label:   f.Label,
			Type:    f.Type,
			Value:   f.Value,
			Purpose: f.Purpose,
		}

		// Associate with section
		if f.Section.ID != "" {
			field.Section = &connect.ItemSection{
				ID:    f.Section.ID,
				Label: f.Section.Label,
			}
		}

		// Include recipe if present
		if f.Recipe != nil {
			field.Recipe = &connect.GeneratorRecipe{
				Length:        f.Recipe.Length,
				CharacterSets: f.Recipe.CharacterSets,
			}
		}

		connectFields = append(connectFields, field)
	}
	return connectFields
}

func toConnectFiles(files []ItemFile) []*connect.File {
	result := make([]*connect.File, 0, len(files))
	for _, f := range files {
		connectFile := &connect.File{
			ID:          f.ID,
			Name:        f.Name,
			Size:        f.Size,
			ContentPath: f.ContentPath,
		}

		// Only set Section if it exists
		if f.Section != nil {
			connectFile.Section = &connect.ItemSection{
				ID:    f.Section.ID,
				Label: f.Section.Label,
			}
		}

		result = append(result, connectFile)
	}
	return result
}
