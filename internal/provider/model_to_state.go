package provider

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func toStateTags(ctx context.Context, modelTags []string, stateTags types.List) (types.List, diag.Diagnostics) {
	var dataTagsSlice []string
	diagnostics := stateTags.ElementsAs(ctx, &dataTagsSlice, false)
	if diagnostics.HasError() {
		return stateTags, diagnostics
	}

	sort.Strings(dataTagsSlice)

	modelTagsSorted := make([]string, len(modelTags))
	copy(modelTagsSorted, modelTags)
	sort.Strings(modelTagsSorted)

	if !reflect.DeepEqual(dataTagsSlice, modelTagsSorted) {
		// If item.Tags is empty, preserve null if the original state was null
		if len(modelTagsSorted) == 0 && stateTags.IsNull() {
			return types.ListNull(types.StringType), nil
		}

		tags, diagnostics := types.ListValueFrom(ctx, types.StringType, modelTagsSorted)
		if diagnostics.HasError() {
			return stateTags, diagnostics
		}
		return tags, nil
	}

	return stateTags, nil
}

func toStateSectionsAndFields(modelSections []model.ItemSection, modelFields []model.ItemField, stateSections []OnePasswordItemResourceSectionModel) []OnePasswordItemResourceSectionModel {
	for _, s := range modelSections {
		section := OnePasswordItemResourceSectionModel{}
		posSection := -1
		newSection := true

		for i := range stateSections {
			existingID := stateSections[i].ID.ValueString()
			existingLabel := stateSections[i].Label.ValueString()
			if (s.ID != "" && s.ID == existingID) || s.Label == existingLabel {
				section = stateSections[i]
				posSection = i
				newSection = false
			}
		}

		section.ID = setStringValue(s.ID)
		section.Label = setStringValuePreservingEmpty(s.Label, section.Label)

		var existingFields []OnePasswordItemResourceFieldModel
		if section.Field != nil {
			existingFields = section.Field
		}
		for _, f := range modelFields {
			if f.SectionID != "" && f.SectionID == s.ID {
				stateField := OnePasswordItemResourceFieldModel{}
				posField := -1
				newField := true

				for i := range existingFields {
					existingID := existingFields[i].ID.ValueString()
					existingLabel := existingFields[i].Label.ValueString()

					if (f.ID != "" && f.ID == existingID) || f.Label == existingLabel {
						stateField = existingFields[i]
						posField = i
						newField = false
					}
				}

				stateField.ID = setStringValue(f.ID)
				stateField.Label = setStringValuePreservingEmpty(f.Label, stateField.Label)
				stateField.Type = setStringValue(string(f.Type))
				stateField.Value = setStringValuePreservingEmpty(f.Value, stateField.Value)

				if f.Recipe != nil {
					charSets := map[string]bool{}
					for _, s := range f.Recipe.CharacterSets {
						charSets[strings.ToLower(string(s))] = true
					}

					stateField.Recipe = []PasswordRecipeModel{{
						Length:  types.Int64Value(int64(f.Recipe.Length)),
						Digits:  types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetDigits))]),
						Symbols: types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetSymbols))]),
					}}
				}

				if newField {
					existingFields = append(existingFields, stateField)
				} else {
					existingFields[posField] = stateField
				}
			}
		}
		section.Field = existingFields

		if newSection {
			stateSections = append(stateSections, section)
		} else {
			stateSections[posSection] = section
		}
	}

	return stateSections
}

func toStateTopLevelFields(modelFields []model.ItemField, state *OnePasswordItemResourceModel) {
	for _, f := range modelFields {
		switch f.Purpose {
		case model.FieldPurposeUsername:
			state.Username = setStringValuePreservingEmpty(f.Value, state.Username)
		case model.FieldPurposePassword:
			state.Password = setStringValue(f.Value)
		case model.FieldPurposeNotes:
			state.NoteValue = setStringValuePreservingEmpty(f.Value, state.NoteValue)
		default:
			if f.SectionID == "" {
				switch f.Label {
				case "username":
					state.Username = setStringValuePreservingEmpty(f.Value, state.Username)
				case "password":
					state.Password = setStringValue(f.Value)
				case "hostname", "server":
					state.Hostname = setStringValuePreservingEmpty(f.Value, state.Hostname)
				case "database":
					state.Database = setStringValuePreservingEmpty(f.Value, state.Database)
				case "port":
					state.Port = setStringValuePreservingEmpty(f.Value, state.Port)
				case "type":
					state.Type = setStringValue(f.Value)
				}
			}
		}
	}
}
