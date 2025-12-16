package provider

import (
	"context"
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func buildLoginFields(data OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
	return []model.ItemField{
		{
			ID:      "username",
			Label:   "username",
			Purpose: model.FieldPurposeUsername,
			Type:    model.FieldTypeString,
			Value:   data.Username.ValueString(),
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
			Value:   data.NoteValue.ValueString(),
		},
	}
}

func buildPasswordFields(data OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
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
			Value:   data.NoteValue.ValueString(),
		},
	}
}

func buildDatabaseFields(data OnePasswordItemResourceModel, password string, recipe *model.GeneratorRecipe) []model.ItemField {
	return []model.ItemField{
		{
			ID:    "username",
			Label: "username",
			Type:  model.FieldTypeString,
			Value: data.Username.ValueString(),
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
			Value: data.Hostname.ValueString(),
		},
		{
			ID:    "database",
			Label: "database",
			Type:  model.FieldTypeString,
			Value: data.Database.ValueString(),
		},
		{
			ID:    "port",
			Label: "port",
			Type:  model.FieldTypeString,
			Value: data.Port.ValueString(),
		},
		{
			ID:    "database_type",
			Label: "type",
			Type:  model.FieldTypeString,
			Value: data.Type.ValueString(),
		},
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   data.NoteValue.ValueString(),
		},
	}
}

func buildSecureNoteFields(data OnePasswordItemResourceModel) []model.ItemField {
	return []model.ItemField{
		{
			ID:      "notesPlain",
			Label:   "notesPlain",
			Type:    model.FieldTypeString,
			Purpose: model.FieldPurposeNotes,
			Value:   data.NoteValue.ValueString(),
		},
	}
}

func buildSectionField(field OnePasswordItemResourceFieldModel, sectionID, sectionLabel string) (*model.ItemField, diag.Diagnostics) {
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

	itemField := &model.ItemField{
		SectionID:    sectionID,
		SectionLabel: sectionLabel,
		ID:           fieldID,
		Type:         model.ItemFieldType(op.ItemFieldType(field.Type.ValueString())),
		Purpose:      model.ItemFieldPurpose(field.Purpose.ValueString()),
		Label:        field.Label.ValueString(),
		Value:        field.Value.ValueString(),
	}

	recipe, err := parseGeneratorRecipe(field.Recipe)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Item conversion error",
			fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
		)}
	}

	if recipe != nil {
		addRecipe(itemField, recipe)
	}

	return itemField, nil
}

func buildSections(data OnePasswordItemResourceModel, item *model.Item) diag.Diagnostics {
	for _, section := range data.Section {
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
		item.Sections = append(item.Sections, s)

		for _, field := range section.Field {
			itemField, diagnostics := buildSectionField(field, s.ID, s.Label)
			if diagnostics.HasError() {
				return diagnostics
			}
			item.Fields = append(item.Fields, *itemField)
		}
	}
	return nil
}

func buildTags(ctx context.Context, data OnePasswordItemResourceModel) ([]string, diag.Diagnostics) {
	var tags []string
	diagnostics := data.Tags.ElementsAs(ctx, &tags, false)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	return tags, nil
}
