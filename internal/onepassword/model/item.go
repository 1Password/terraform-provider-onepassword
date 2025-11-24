package model

import (
	"context"
	"fmt"

	connect "github.com/1Password/connect-sdk-go/onepassword"
	sdk "github.com/1password/onepassword-sdk-go"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
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
	i.Category = fromSDKCategoryToModel(item.Category)
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
		Category: fromModelCategoryToSDK(i.Category),
		Tags:     i.Tags,
		Sections: toSDKSections(i.Sections),
		Websites: toSDKWebsites(i.URLs),
	}

	params.Fields, params.Notes = toSDKFields(i.Fields)

	return params
}

// FromConnectItemToModel creates a new Item from a Connect SDK item
func (i *Item) FromConnectItemToModel(item *connect.Item) error {
	if item == nil {
		return fmt.Errorf("cannot convert nil Connect item to model")
	}

	i.ID = item.ID
	i.Title = item.Title
	i.VaultID = item.Vault.ID
	i.Category = ItemCategory(item.Category)
	i.Version = item.Version
	i.Tags = item.Tags
	i.URLs = fromConnectURLs(item.URLs)

	// Convert sections/fields/files
	sectionMap := make(map[string]ItemSection)
	i.Sections = fromConnectSections(item.Sections, sectionMap)
	i.Files = fromConnectFiles(item.Files, sectionMap)

	fields, err := fromConnectFields(item.Fields, sectionMap)
	if err != nil {
		return err
	}
	i.Fields = fields

	return nil
}

// FromModelItemToConnect creates a Connect SDK item from a model Item
func (i *Item) FromModelItemToConnect() (*connect.Item, error) {
	fields, err := toConnectFields(i.Fields)
	if err != nil {
		return nil, err
	}

	return &connect.Item{
		ID:       i.ID,
		Title:    i.Title,
		Vault:    connect.ItemVault{ID: i.VaultID},
		Category: connect.ItemCategory(i.Category),
		Version:  i.Version,
		Tags:     i.Tags,
		URLs:     toConnectURLs(i.URLs),
		Sections: toConnectSections(i.Sections),
		Fields:   fields,
		Files:    toConnectFiles(i.Files),
	}, nil
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
			Type:  toModelFieldType(f.FieldType),
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
		FieldType: toSDKFieldType(f.Type),
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

func fromConnectFields(fields []*connect.ItemField, sectionMap map[string]ItemSection) ([]ItemField, error) {
	modelFields := make([]ItemField, 0, len(fields))
	for _, f := range fields {
		if f == nil {
			continue
		}

		field := ItemField{
			ID:      f.ID,
			Label:   f.Label,
			Type:    ItemFieldType(f.Type),
			Value:   f.Value,
			Purpose: ItemFieldPurpose(f.Purpose),
		}

		// Provider handles dates in `YYYY-MM-DD` format.
		// Connect returns dates as timestamp
		// Converting timestamp to `YYYY-MM-DD` string.
		if f.Type == connect.FieldTypeDate {
			dateStr, err := util.SecondsToYYYYMMDD(field.Value)
			if err != nil {
				return modelFields, fmt.Errorf("fromConnectFields: failed to parse timestamp %s to 'YYYY-MM-DD' string format: %w", field.Value, err)
			}
			field.Value = dateStr
		}

		// Associate field with section if applicable
		if f.Section != nil && f.Section.ID != "" {
			if section, exists := sectionMap[f.Section.ID]; exists {
				field.SectionID = section.ID
				field.SectionLabel = section.Label
			}
		}

		modelFields = append(modelFields, field)
	}
	return modelFields, nil
}

func fromConnectFiles(files []*connect.File, sectionMap map[string]ItemSection) []ItemFile {
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
		if f.Section != nil && f.Section.ID != "" {
			if section, exists := sectionMap[f.Section.ID]; exists {
				itemFile.SectionID = section.ID
				itemFile.SectionLabel = section.Label
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

func toConnectFields(fields []ItemField) ([]*connect.ItemField, error) {
	connectFields := make([]*connect.ItemField, 0, len(fields))
	for _, f := range fields {
		field := &connect.ItemField{
			ID:       f.ID,
			Label:    f.Label,
			Value:    f.Value,
			Generate: f.Generate,
			Type:     connect.ItemFieldType(f.Type),
			Purpose:  connect.ItemFieldPurpose(f.Purpose),
		}

		if field.Type == connect.FieldTypeDate {
			// Convert date string to timestamp to bypass Connect's timezone-dependent parsing
			// and ensure consistent storage regardless of where Connect is deployed.
			timestamp, err := util.YYYYMMDDToSeconds(field.Value)
			if err != nil {
				return connectFields, fmt.Errorf("toConnectFields: failed to convert '%s' date string to timestamp: %w", field.Value, err)
			}
			field.Value = timestamp
		}

		// Associate with section
		if f.SectionID != "" {
			field.Section = &connect.ItemSection{
				ID:    f.SectionID,
				Label: f.SectionLabel,
			}
		}

		// Include recipe if present
		if f.Recipe != nil {
			// Connect allows confiugration of letters for password recipes
			// We need to include letters in the character sets in order to ensure they are not excluded
			characterSets := []string{"LETTERS"}

			for _, cs := range f.Recipe.CharacterSets {
				characterSets = append(characterSets, string(cs))
			}

			field.Recipe = &connect.GeneratorRecipe{
				Length:        f.Recipe.Length,
				CharacterSets: characterSets,
			}
		}

		connectFields = append(connectFields, field)

	}
	return connectFields, nil
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
		if f.SectionID != "" {
			connectFile.Section = &connect.ItemSection{
				ID:    f.SectionID,
				Label: f.SectionLabel,
			}
		}

		result = append(result, connectFile)
	}
	return result
}

var modelToSdkFiledTypeMap = map[ItemFieldType]sdk.ItemFieldType{
	FieldTypeConcealed: sdk.ItemFieldTypeConcealed,
	FieldTypeDate:      sdk.ItemFieldTypeDate,
	FieldTypeEmail:     sdk.ItemFieldTypeEmail,
	FieldTypeMenu:      sdk.ItemFieldTypeMenu,
	FieldTypeMonthYear: sdk.ItemFieldTypeMonthYear,
	FieldTypeOTP:       sdk.ItemFieldTypeTOTP,
	FieldTypeString:    sdk.ItemFieldTypeText,
	FieldTypeURL:       sdk.ItemFieldTypeURL,
}

func toSDKFieldType(filedType ItemFieldType) sdk.ItemFieldType {
	return modelToSdkFiledTypeMap[filedType]
}

var sdkToModelFieldTypeMap = map[sdk.ItemFieldType]ItemFieldType{
	sdk.ItemFieldTypeConcealed: FieldTypeConcealed,
	sdk.ItemFieldTypeDate:      FieldTypeDate,
	sdk.ItemFieldTypeEmail:     FieldTypeEmail,
	sdk.ItemFieldTypeMenu:      FieldTypeMenu,
	sdk.ItemFieldTypeMonthYear: FieldTypeMonthYear,
	sdk.ItemFieldTypeTOTP:      FieldTypeOTP,
	sdk.ItemFieldTypeText:      FieldTypeString,
	sdk.ItemFieldTypeURL:       FieldTypeURL,
}

func toModelFieldType(filedType sdk.ItemFieldType) ItemFieldType {
	return sdkToModelFieldTypeMap[filedType]
}

var modelToSDKCategoryMap = map[ItemCategory]sdk.ItemCategory{
	Login:      sdk.ItemCategoryLogin,
	Password:   sdk.ItemCategoryPassword,
	SecureNote: sdk.ItemCategorySecureNote,
	Document:   sdk.ItemCategoryDocument,
	SSHKey:     sdk.ItemCategorySSHKey,
	Database:   sdk.ItemCategoryDatabase,
}

func fromModelCategoryToSDK(itemCategory ItemCategory) sdk.ItemCategory {
	return modelToSDKCategoryMap[itemCategory]
}

var sdkToModelCategoryMap = map[sdk.ItemCategory]ItemCategory{
	sdk.ItemCategoryLogin:      Login,
	sdk.ItemCategoryPassword:   Password,
	sdk.ItemCategorySecureNote: SecureNote,
	sdk.ItemCategoryDocument:   Document,
	sdk.ItemCategorySSHKey:     SSHKey,
	sdk.ItemCategoryDatabase:   Database,
}

func fromSDKCategoryToModel(itemCategory sdk.ItemCategory) ItemCategory {
	return sdkToModelCategoryMap[itemCategory]
}
