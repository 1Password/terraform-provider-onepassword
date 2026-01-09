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

func toStateSectionsAndFieldsList(modelSections []model.ItemSection, modelFields []model.ItemField, stateSections []OnePasswordItemResourceSectionListModel) []OnePasswordItemResourceSectionListModel {
	for _, s := range modelSections {
		section := OnePasswordItemResourceSectionListModel{}
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
		if section.FieldList != nil {
			existingFields = section.FieldList
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
		section.FieldList = existingFields

		if newSection {
			stateSections = append(stateSections, section)
		} else {
			stateSections[posSection] = section
		}
	}

	return stateSections
}

func toStateSectionsAndFieldsMap(item *model.Item, stateSectionMap map[string]OnePasswordItemResourceSectionMapModel) map[string]OnePasswordItemResourceSectionMapModel {
	sectionMap := make(map[string]OnePasswordItemResourceSectionMapModel)

	for _, modelSection := range item.Sections {
		section := OnePasswordItemResourceSectionMapModel{
			ID:       types.StringValue(modelSection.ID),
			FieldMap: make(map[string]OnePasswordItemResourceFieldMapModel),
		}

		for _, modelField := range item.Fields {
			// Only process fields that belong to this section
			if modelField.SectionID != modelSection.ID {
				continue
			}

			field := OnePasswordItemResourceFieldMapModel{
				ID:   setStringValue(modelField.ID),
				Type: setStringValue(string(modelField.Type)),
			}

			existingSection, sectionExists := stateSectionMap[modelSection.Label]
			if sectionExists {
				if existingField, fieldExists := existingSection.FieldMap[modelField.Label]; fieldExists {
					field.Value = setStringValuePreservingEmpty(modelField.Value, existingField.Value)
				} else {
					field.Value = setStringValuePreservingEmpty(modelField.Value, types.StringNull())
				}
			} else {
				field.Value = setStringValuePreservingEmpty(modelField.Value, types.StringNull())
			}

			if modelField.Recipe != nil {
				charSets := map[string]bool{}
				for _, s := range modelField.Recipe.CharacterSets {
					charSets[strings.ToLower(string(s))] = true
				}

				field.Recipe = &PasswordRecipeModel{
					Length:  types.Int64Value(int64(modelField.Recipe.Length)),
					Digits:  types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetDigits))]),
					Symbols: types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetSymbols))]),
				}
			} else if sectionExists {
				// If server didn't return a recipe - preserve from existing plan/state if available
				if existingField, fieldExists := existingSection.FieldMap[modelField.Label]; fieldExists {
					field.Recipe = existingField.Recipe
				}
			}

			section.FieldMap[modelField.Label] = field
		}

		sectionMap[modelSection.Label] = section
	}

	return sectionMap
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
