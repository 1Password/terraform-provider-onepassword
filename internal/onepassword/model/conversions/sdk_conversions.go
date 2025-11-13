package conversions

import (
	"context"
	"fmt"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	sdk "github.com/1password/onepassword-sdk-go"
	"github.com/hashicorp/go-uuid"
)

// ToSDKItem converts internal model Item to SDK ItemCreateParams
func ToSDKItem(i *model.Item, vaultID string) sdk.ItemCreateParams {
	params := sdk.ItemCreateParams{
		VaultID:  vaultID,
		Title:    i.Title,
		Category: toSDKCategory(i.Category),
		Tags:     toSDKTags(i.Tags),
	}

	// Convert fields
	for _, field := range i.Fields {
		if field.Purpose == model.FieldPurposeNotes {
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

func toSDKFieldType(modelType model.ItemFieldType) sdk.ItemFieldType {
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

func toSDKCategory(modelType model.ItemCategory) sdk.ItemCategory {
	switch modelType {
	case model.ItemCategorySecureNote:
		return "SecureNote"
	case model.ItemCategorySSHKey:
		return "SshKey"
	default:
		return sdk.ItemCategory(modelType)
	}
}

func toSDKField(f *model.ItemField) sdk.ItemField {
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
		FieldType: toSDKFieldType(f.Type),
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

func generatePassword(recipe *model.GeneratorRecipe) (string, error) {
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
