package provider

import (
	"context"
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
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

	parsed := &model.GeneratorRecipe{
		Length:        32,
		CharacterSets: []model.CharacterSet{},
	}

	length := recipe.Length.ValueInt64()
	if length > 64 {
		return nil, fmt.Errorf("password_recipe.length must be an integer between 1 and 64")
	}

	if length > 0 {
		parsed.Length = int(length)
	}

	if recipe.Digits.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetDigits)
	}
	if recipe.Symbols.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetSymbols)
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
