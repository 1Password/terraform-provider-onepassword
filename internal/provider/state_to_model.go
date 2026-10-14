package provider

import (
	"context"
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	opssh "github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/ssh"
)

func toModelLoginFields(state OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
	return []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   state.Username.ValueString(),
		},
		{
			ID:       "password",
			Label:    "password",
			Purpose:  model.FieldPurposePassword,
			Type:     model.FieldTypeConcealed,
			Value:    password,
			Generate: password == "",
			Recipe:   recipe,
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

func toModelPasswordFields(state OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
	return []model.ItemField{
		{
			ID:       "password",
			Label:    "password",
			Purpose:  model.FieldPurposePassword,
			Type:     model.FieldTypeConcealed,
			Value:    password,
			Generate: password == "",
			Recipe:   recipe,
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

func toModelDatabaseFields(state OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
	return []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Type:  model.FieldTypeString,
			Value: state.Username.ValueString(),
		},
		{
			ID:       "password",
			Label:    "password",
			Type:     model.FieldTypeConcealed,
			Value:    password,
			Generate: password == "",
			Recipe:   recipe,
		},
		{
			ID:    "hostname",
			Label: "hostname",
			Type:  model.FieldTypeString,
			Value: state.Hostname.ValueString(),
		},
		{
			ID:    "database",
			Label: "database",
			Type:  model.FieldTypeString,
			Value: state.Database.ValueString(),
		},
		{
			ID:    "port",
			Label: "port",
			Type:  model.FieldTypeString,
			Value: state.Port.ValueString(),
		},
		{
			ID:    "database_type",
			Label: "type",
			Type:  model.FieldTypeString,
			Value: state.Type.ValueString(),
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

func toModelSecureNoteFields(state OnePasswordItemResourceModel) []model.ItemField {
	return []model.ItemField{
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

// toModelSSHKeyFields builds the fields for an SSH key item. On create the
// key pair is generated; on update the existing private key from state (in
// OpenSSH form) is converted back to the PKCS#8 form 1Password stores. Only
// the private key is written: the account API synthesizes the public key,
// fingerprint, and key type from it.
func toModelSSHKeyFields(state OnePasswordItemResourceModel) ([]model.ItemField, diag.Diagnostics) {
	var privateKeyPKCS8 string

	if state.PrivateKey.ValueString() == "" {
		key, err := opssh.GenerateKeyPair(state.SSHKeyType.ValueString(), int(state.SSHKeyBits.ValueInt64()))
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"SSH key generation error",
				fmt.Sprintf("Failed to generate SSH key: %v", err),
			)}
		}
		privateKeyPKCS8 = key.PrivateKeyPKCS8
	} else {
		pkcs8, err := opssh.OpenSSHToPKCS8(state.PrivateKey.ValueString())
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"SSH key conversion error",
				fmt.Sprintf("Failed to convert private key back to PKCS#8: %v", err),
			)}
		}
		privateKeyPKCS8 = pkcs8
	}

	fields := []model.ItemField{
		{
			ID:    "private_key",
			Label: "private key",
			Type:  model.FieldTypeSSHKey,
			Value: privateKeyPKCS8,
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}

	return fields, nil
}

func toModelAPICredentialFields(state OnePasswordItemResourceModel) []model.ItemField {
	return []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   state.Username.ValueString(),
		},
		{
			ID:    "credential",
			Label: "credential",
			Type:  model.FieldTypeConcealed,
			Value: state.Credential.ValueString(),
		},
		{
			ID:    "validFrom",
			Label: "valid from",
			Type:  model.FieldTypeDate,
			Value: state.ValidFrom.ValueString(),
		},
		{
			ID:    "filename",
			Label: "filename",
			Type:  model.FieldTypeString,
			Value: state.Filename.ValueString(),
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

func toModelServerFields(state OnePasswordItemResourceModel) ([]model.ItemField, []model.ItemSection) {
	fields := []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   state.Username.ValueString(),
		},
		{
			ID:       "password",
			Label:    "password",
			Purpose:  model.FieldPurposePassword,
			Type:     model.FieldTypeConcealed,
			Value:    state.Password.ValueString(),
			Generate: state.Password.ValueString() == "",
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}

	var sections []model.ItemSection

	if state.AdminConsoleURL.ValueString() != "" ||
		state.AdminConsoleUsername.ValueString() != "" ||
		state.AdminConsolePassword.ValueString() != "" {
		sectionID := "admin_console"
		sectionLabel := "Admin Console"

		sections = append(sections, model.ItemSection{ID: sectionID, Label: sectionLabel})

		adminFields := []model.ItemField{
			{
				ID:           "admin_console_url",
				Label:        "admin console URL",
				Type:         model.FieldTypeString,
				Value:        state.AdminConsoleURL.ValueString(),
				SectionID:    sectionID,
				SectionLabel: sectionLabel,
			},
			{
				ID:           "admin_console_username",
				Label:        "admin console username",
				Type:         model.FieldTypeString,
				Value:        state.AdminConsoleUsername.ValueString(),
				SectionID:    sectionID,
				SectionLabel: sectionLabel,
			},
			{
				ID:           "admin_console_password",
				Label:        "console password",
				Type:         model.FieldTypeConcealed,
				Value:        state.AdminConsolePassword.ValueString(),
				SectionID:    sectionID,
				SectionLabel: sectionLabel,
			},
		}
		fields = append(fields, adminFields...)
	}

	return fields, sections
}

func toModelRouterFields(state OnePasswordItemResourceModel) []model.ItemField {
	return []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Type:  model.FieldTypeString,
			Value: state.Username.ValueString(),
		},
		{
			ID:       "password",
			Label:    "base station password",
			Purpose:  model.FieldPurposePassword,
			Type:     model.FieldTypeConcealed,
			Value:    state.Password.ValueString(),
			Generate: state.Password.ValueString() == "",
		},
		{
			ID:    "server",
			Label: "server / IP address",
			Type:  model.FieldTypeString,
			Value: state.ServerAddress.ValueString(),
		},
		{
			ID:    "network_name",
			Label: "network name",
			Type:  model.FieldTypeString,
			Value: state.NetworkName.ValueString(),
		},
		{
			ID:    "wireless_security",
			Label: "wireless security",
			Type:  model.FieldTypeMenu,
			Value: state.WirelessSecurity.ValueString(),
		},
		{
			ID:    "wireless_password",
			Label: "wireless network password",
			Type:  model.FieldTypeConcealed,
			Value: state.WirelessPassword.ValueString(),
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}
}

func toModelSoftwareLicenseFields(state OnePasswordItemResourceModel) ([]model.ItemField, []model.ItemSection) {
	fields := []model.ItemField{
		{
			ID:    "reg_code",
			Label: "license key",
			Type:  model.FieldTypeString,
			Value: state.LicenseKey.ValueString(),
		},
		{
			ID:    "product_version",
			Label: "version",
			Type:  model.FieldTypeString,
			Value: state.Version.ValueString(),
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   state.NoteValue.ValueString(),
		},
	}

	var sections []model.ItemSection

	if state.LicensedTo.ValueString() != "" || state.RegisteredEmail.ValueString() != "" {
		sectionID := "customer"
		sectionLabel := "Customer"

		sections = append(sections, model.ItemSection{ID: sectionID, Label: sectionLabel})

		fields = append(fields, model.ItemField{
			ID:           "reg_name",
			Label:        "licensed to",
			Type:         model.FieldTypeString,
			Value:        state.LicensedTo.ValueString(),
			SectionID:    sectionID,
			SectionLabel: sectionLabel,
		})
		fields = append(fields, model.ItemField{
			ID:           "reg_email",
			Label:        "registered email",
			Type:         model.FieldTypeEmail,
			Value:        state.RegisteredEmail.ValueString(),
			SectionID:    sectionID,
			SectionLabel: sectionLabel,
		})
	}

	if state.DownloadLink.ValueString() != "" {
		sectionID := "publisher"
		sectionLabel := "Publisher"

		sections = append(sections, model.ItemSection{ID: sectionID, Label: sectionLabel})

		fields = append(fields, model.ItemField{
			ID:           "download_link",
			Label:        "download page",
			Type:         model.FieldTypeURL,
			Value:        state.DownloadLink.ValueString(),
			SectionID:    sectionID,
			SectionLabel: sectionLabel,
		})
	}

	return fields, sections
}

func toModelSectionField(field OnePasswordItemResourceFieldModel, sectionID, sectionLabel string) (*model.ItemField, diag.Diagnostics) {
	fieldID := field.ID.ValueString()
	// Generate field ID if empty
	if fieldID == "" {
		sid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Item conversion error",
				fmt.Sprintf("Unable to generate a field ID, has error: %v", err),
			)}
		}
		fieldID = sid
	}

	modelItemField := &model.ItemField{
		SectionID:    sectionID,
		SectionLabel: sectionLabel,
		ID:           fieldID,
		Type:         model.ItemFieldType(op.ItemFieldType(field.Type.ValueString())),
		Label:        field.Label.ValueString(),
		Value:        field.Value.ValueString(),
	}

	recipe, err := parseGeneratorRecipeList(field.Recipe)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Item conversion error",
			fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
		)}
	}

	if recipe != nil {
		addRecipe(modelItemField, recipe)
	}

	return modelItemField, nil
}

func toModelSections(state OnePasswordItemResourceModel, modelItem *model.Item) diag.Diagnostics {
	for _, section := range state.SectionList {
		sectionID := section.ID.ValueString()
		if sectionID == "" {
			sid, err := uuid.GenerateUUID()
			if err != nil {
				return diag.Diagnostics{diag.NewErrorDiagnostic(
					"Item conversion error",
					fmt.Sprintf("Unable to generate a section ID, has error: %v", err),
				)}
			}
			sectionID = sid
		}

		s := model.ItemSection{
			ID:    sectionID,
			Label: section.Label.ValueString(),
		}
		modelItem.Sections = append(modelItem.Sections, s)

		for _, field := range section.FieldList {
			modelItemField, diagnostics := toModelSectionField(field, s.ID, s.Label)
			if diagnostics.HasError() {
				return diagnostics
			}
			modelItem.Fields = append(modelItem.Fields, *modelItemField)
		}
	}
	return nil
}

func toModelTags(ctx context.Context, state OnePasswordItemResourceModel) ([]string, diag.Diagnostics) {
	var tags []string
	diagnostics := state.Tags.ElementsAs(ctx, &tags, false)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	return tags, nil
}

func parseGeneratorRecipeFromModel(recipe *PasswordRecipeModel) (*model.GeneratorRecipe, error) {
	if recipe == nil {
		return nil, nil
	}

	kind := model.RecipeKindRandom
	switch recipe.Type.ValueString() {
	case "memorable":
		kind = model.RecipeKindMemorable
	case "pin":
		kind = model.RecipeKindPin
	case "", "random":
	default:
		return nil, fmt.Errorf("password_recipe.type must be one of %v", recipeTypes)
	}

	parsed := &model.GeneratorRecipe{
		Kind:              kind,
		CharacterSets:     []model.CharacterSet{},
		ExcludeCharacters: recipe.ExcludeCharacters.ValueString(),
		Separator:         recipe.Separator.ValueString(),
		WordList:          recipe.WordList.ValueString(),
		Capitalize:        recipe.Capitalize.ValueBool(),
	}

	length := recipe.Length.ValueInt64()
	if length > 64 {
		return nil, fmt.Errorf("password_recipe.length must be an integer between 1 and 64")
	}

	if length > 0 {
		parsed.Length = int(length)
	} else {
		parsed.Length = 32
	}

	wordCount := recipe.WordCount.ValueInt64()
	if wordCount != 0 && (wordCount < 3 || wordCount > 15) {
		return nil, fmt.Errorf("password_recipe.word_count must be an integer between 3 and 15")
	}
	if wordCount > 0 {
		parsed.WordCount = int(wordCount)
	} else {
		parsed.WordCount = 3
	}

	if kind == model.RecipeKindRandom {
		if recipe.Digits.ValueBool() {
			parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetDigits)
		}
		if recipe.Symbols.ValueBool() {
			parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetSymbols)
		}
	}

	return parsed, nil
}

func toModelSectionFieldMap(field OnePasswordItemResourceFieldMapModel, fieldLabel, sectionID, sectionLabel string) (*model.ItemField, diag.Diagnostics) {
	fieldID := field.ID.ValueString()
	// Generate field ID if empty
	if fieldID == "" {
		sid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Item conversion error",
				fmt.Sprintf("Unable to generate a field ID, has error: %v", err),
			)}
		}
		fieldID = sid
	}

	modelItemField := &model.ItemField{
		SectionID:    sectionID,
		SectionLabel: sectionLabel,
		ID:           fieldID,
		Type:         model.ItemFieldType(op.ItemFieldType(field.Type.ValueString())),
		Label:        fieldLabel,
		Value:        field.Value.ValueString(),
	}

	recipe, err := parseGeneratorRecipeFromModel(field.Recipe)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Item conversion error",
			fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
		)}
	}

	if recipe != nil {
		addRecipe(modelItemField, recipe)
	}

	return modelItemField, nil
}

func toModelSectionsFromMap(state OnePasswordItemResourceModel, modelItem *model.Item) diag.Diagnostics {
	for sectionLabel, section := range state.SectionMap {
		sectionID := section.ID.ValueString()
		if sectionID == "" {
			sid, err := uuid.GenerateUUID()
			if err != nil {
				return diag.Diagnostics{diag.NewErrorDiagnostic(
					"Item conversion error",
					fmt.Sprintf("Unable to generate a section ID, has error: %v", err),
				)}
			}
			sectionID = sid
		}

		s := model.ItemSection{
			ID:    sectionID,
			Label: sectionLabel, // Use the map key as the label
		}
		modelItem.Sections = append(modelItem.Sections, s)

		for fieldLabel, field := range section.FieldMap {
			modelItemField, diagnostics := toModelSectionFieldMap(field, fieldLabel, s.ID, s.Label)
			if diagnostics.HasError() {
				return diagnostics
			}
			modelItem.Fields = append(modelItem.Fields, *modelItemField)
		}
	}
	return nil
}
