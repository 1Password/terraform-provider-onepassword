package provider

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	opssh "github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/ssh"
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

func toStateRecipe(r *model.GeneratorRecipe) PasswordRecipeModel {
	kind := "random"
	switch r.Kind {
	case model.RecipeKindMemorable:
		kind = "memorable"
	case model.RecipeKindPin:
		kind = "pin"
	}

	charSets := map[string]bool{}
	for _, s := range r.CharacterSets {
		charSets[strings.ToLower(string(s))] = true
	}

	separator := r.Separator
	if separator == "" {
		separator = "hyphens"
	}
	wordList := r.WordList
	if wordList == "" {
		wordList = "full_words"
	}
	wordCount := r.WordCount
	if wordCount == 0 {
		wordCount = 3
	}

	return PasswordRecipeModel{
		Type:              types.StringValue(kind),
		Length:            types.Int64Value(int64(r.Length)),
		Digits:            types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetDigits))]),
		Symbols:           types.BoolValue(charSets[strings.ToLower(string(model.CharacterSetSymbols))]),
		ExcludeCharacters: setStringValue(r.ExcludeCharacters),
		WordCount:         types.Int64Value(int64(wordCount)),
		Separator:         types.StringValue(separator),
		Capitalize:        types.BoolValue(r.Capitalize),
		WordList:          types.StringValue(wordList),
	}
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
					recipe := toStateRecipe(f.Recipe)
					stateField.Recipe = []PasswordRecipeModel{recipe}
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
				recipe := toStateRecipe(modelField.Recipe)
				field.Recipe = &recipe
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

// categoryManagedSectionIDs lists the 1Password template section IDs that the
// provider itself creates for a category. They are excluded from the generic
// section state because their fields surface as top-level attributes; showing
// them as sections would create permanent drift against user configuration.
func categoryManagedSectionIDs(category model.ItemCategory) map[string]bool {
	switch category {
	case model.Server:
		return map[string]bool{"admin_console": true}
	case model.SoftwareLicense:
		return map[string]bool{"customer": true, "publisher": true}
	default:
		return nil
	}
}

// toStateCategoryFields maps category-specific fields back to their top-level
// attributes, matching by field ID regardless of section.
func toStateCategoryFields(modelItem *model.Item, state *OnePasswordItemResourceModel) {
	byID := func(id string) (model.ItemField, bool) {
		for _, f := range modelItem.Fields {
			if f.ID == id {
				return f, true
			}
		}
		return model.ItemField{}, false
	}

	switch modelItem.Category {
	case model.SSHKey:
		if f, ok := byID("private_key"); ok && f.Value != "" {
			if openSSH, err := opssh.PrivateKeyToOpenSSH([]byte(f.Value), modelItem.ID); err == nil {
				state.PrivateKey = setStringValue(openSSH)
			} else {
				state.PrivateKey = setStringValue(f.Value)
			}

			// The public key, fingerprint, and key type are derived from the
			// stored private key: the account API synthesizes them for
			// SDK-created items, while Connect items may not carry them at
			// all (see 1Password/connect#107).
			if pub, err := opssh.PublicKeyFromPrivateKey(f.Value); err == nil {
				state.PublicKey = setStringValue(pub)
				if fingerprint, err := opssh.PublicKeyFingerprint(pub); err == nil {
					state.Fingerprint = setStringValue(fingerprint)
				}
				if keyType, err := opssh.KeyTypeFromPublicKey(pub); err == nil {
					state.SSHKeyTypeOf = setStringValue(keyType)

					// ssh_key_type and ssh_key_bits are not stored on the item;
					// derive them from the public key so imported state matches.
					if state.SSHKeyType.IsNull() || state.SSHKeyType.IsUnknown() {
						switch {
						case keyType == "Ed25519":
							state.SSHKeyType = types.StringValue("ed25519")
							if state.SSHKeyBits.IsNull() || state.SSHKeyBits.IsUnknown() {
								state.SSHKeyBits = types.Int64Value(2048)
							}
						case strings.HasPrefix(keyType, "RSA, "):
							state.SSHKeyType = types.StringValue("rsa")
							bitsStr := strings.TrimSuffix(strings.TrimPrefix(keyType, "RSA, "), "-bit")
							if bits, err := strconv.Atoi(bitsStr); err == nil && (state.SSHKeyBits.IsNull() || state.SSHKeyBits.IsUnknown()) {
								state.SSHKeyBits = types.Int64Value(int64(bits))
							}
						}
					}
				}
			}
		}
	case model.APICredential:
		if f, ok := byID("credential"); ok {
			state.Credential = setStringValuePreservingEmpty(f.Value, state.Credential)
		}
		if f, ok := byID("validFrom"); ok {
			state.ValidFrom = setStringValuePreservingEmpty(f.Value, state.ValidFrom)
		}
		if f, ok := byID("filename"); ok {
			state.Filename = setStringValuePreservingEmpty(f.Value, state.Filename)
		}
	case model.Server:
		if f, ok := byID("admin_console_url"); ok {
			state.AdminConsoleURL = setStringValuePreservingEmpty(f.Value, state.AdminConsoleURL)
		}
		if f, ok := byID("admin_console_username"); ok {
			state.AdminConsoleUsername = setStringValuePreservingEmpty(f.Value, state.AdminConsoleUsername)
		}
		if f, ok := byID("admin_console_password"); ok {
			state.AdminConsolePassword = setStringValuePreservingEmpty(f.Value, state.AdminConsolePassword)
		}
	case model.Router:
		if f, ok := byID("network_name"); ok {
			state.NetworkName = setStringValuePreservingEmpty(f.Value, state.NetworkName)
		}
		if f, ok := byID("server"); ok {
			state.ServerAddress = setStringValuePreservingEmpty(f.Value, state.ServerAddress)
		}
		if f, ok := byID("wireless_security"); ok {
			state.WirelessSecurity = setStringValuePreservingEmpty(f.Value, state.WirelessSecurity)
		}
		if f, ok := byID("wireless_password"); ok {
			state.WirelessPassword = setStringValuePreservingEmpty(f.Value, state.WirelessPassword)
		}
	case model.SoftwareLicense:
		if f, ok := byID("reg_code"); ok {
			state.LicenseKey = setStringValuePreservingEmpty(f.Value, state.LicenseKey)
		}
		if f, ok := byID("product_version"); ok {
			state.Version = setStringValuePreservingEmpty(f.Value, state.Version)
		}
		if f, ok := byID("download_link"); ok {
			state.DownloadLink = setStringValuePreservingEmpty(f.Value, state.DownloadLink)
		}
		if f, ok := byID("reg_name"); ok {
			state.LicensedTo = setStringValuePreservingEmpty(f.Value, state.LicensedTo)
		}
		if f, ok := byID("reg_email"); ok {
			state.RegisteredEmail = setStringValuePreservingEmpty(f.Value, state.RegisteredEmail)
		}
	}
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
