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

func processTags(ctx context.Context, itemTags []string, currentTags types.List) (types.List, diag.Diagnostics) {
	var dataTagsSlice []string
	diagnostics := currentTags.ElementsAs(ctx, &dataTagsSlice, false)
	if diagnostics.HasError() {
		return currentTags, diagnostics
	}

	sort.Strings(dataTagsSlice)
	if !reflect.DeepEqual(dataTagsSlice, itemTags) {
		// If item.Tags is empty, preserve null if the original state was null
		if len(itemTags) == 0 && currentTags.IsNull() {
			return types.ListNull(types.StringType), nil
		}

		tags, diagnostics := types.ListValueFrom(ctx, types.StringType, itemTags)
		if diagnostics.HasError() {
			return currentTags, diagnostics
		}
		return tags, nil
	}

	return currentTags, nil
}

func processSectionsAndFields(itemSections []model.ItemSection, itemFields []model.ItemField, dataSections []OnePasswordItemResourceSectionModel) []OnePasswordItemResourceSectionModel {
	for _, s := range itemSections {
		section := OnePasswordItemResourceSectionModel{}
		posSection := -1
		newSection := true

		for i := range dataSections {
			existingID := dataSections[i].ID.ValueString()
			existingLabel := dataSections[i].Label.ValueString()
			if (s.ID != "" && s.ID == existingID) || s.Label == existingLabel {
				section = dataSections[i]
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
		for _, f := range itemFields {
			if f.SectionID != "" && f.SectionID == s.ID {
				dataField := OnePasswordItemResourceFieldModel{}
				posField := -1
				newField := true

				for i := range existingFields {
					existingID := existingFields[i].ID.ValueString()
					existingLabel := existingFields[i].Label.ValueString()

					if (f.ID != "" && f.ID == existingID) || f.Label == existingLabel {
						dataField = existingFields[i]
						posField = i
						newField = false
					}
				}

				dataField.ID = setStringValue(f.ID)
				dataField.Label = setStringValuePreservingEmpty(f.Label, dataField.Label)
				dataField.Purpose = setStringValue(string(f.Purpose))
				dataField.Type = setStringValue(string(f.Type))
				dataField.Value = setStringValuePreservingEmpty(f.Value, dataField.Value)

				if f.Recipe != nil {
					charSets := map[string]bool{}
					for _, s := range f.Recipe.CharacterSets {
						charSets[strings.ToLower(string(s))] = true
					}

					dataField.Recipe = []PasswordRecipeModel{{
						Length:  types.Int64Value(int64(f.Recipe.Length)),
						Digits:  types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetDigits))]),
						Symbols: types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetSymbols))]),
					}}
				}

				if newField {
					existingFields = append(existingFields, dataField)
				} else {
					existingFields[posField] = dataField
				}
			}
		}
		section.Field = existingFields

		if newSection {
			dataSections = append(dataSections, section)
		} else {
			dataSections[posSection] = section
		}
	}

	return dataSections
}

func processTopLevelFields(itemFields []model.ItemField, data *OnePasswordItemResourceModel) {
	for _, f := range itemFields {
		switch f.Purpose {
		case model.FieldPurposeUsername:
			data.Username = setStringValuePreservingEmpty(f.Value, data.Username)
		case model.FieldPurposePassword:
			data.Password = setStringValue(f.Value)
		case model.FieldPurposeNotes:
			data.NoteValue = setStringValuePreservingEmpty(f.Value, data.NoteValue)
		default:
			if f.SectionID == "" {
				switch f.Label {
				case "username":
					data.Username = setStringValuePreservingEmpty(f.Value, data.Username)
				case "password":
					data.Password = setStringValue(f.Value)
				case "hostname", "server":
					data.Hostname = setStringValuePreservingEmpty(f.Value, data.Hostname)
				case "database":
					data.Database = setStringValuePreservingEmpty(f.Value, data.Database)
				case "port":
					data.Port = setStringValuePreservingEmpty(f.Value, data.Port)
				case "type":
					data.Type = setStringValue(f.Value)
				}
			}
		}
	}
}
