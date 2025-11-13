package model

import (
	"context"
	"fmt"

	sdk "github.com/1password/onepassword-sdk-go"
	"github.com/hashicorp/go-uuid"
)

func toSDKFieldType(modelType ItemFieldType) sdk.ItemFieldType {
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

func toSDKCategory(modelType ItemCategory) sdk.ItemCategory {
	switch modelType {
	case ItemCategorySecureNote:
		return "SecureNote"
	case ItemCategorySSHKey:
		return "SshKey"
	default:
		return sdk.ItemCategory(modelType)
	}
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
