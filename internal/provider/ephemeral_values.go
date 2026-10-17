package provider

import (
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// handleWriteOnlyField sets a plan field from its write-only counterpart if the version is set
func handleWriteOnlyField(version types.Int64, woValue types.String, planValue *types.String) {
	if !version.IsNull() {
		if !woValue.IsNull() && !woValue.IsUnknown() {
			*planValue = woValue
		}
	}
}

// clearWriteOnlyFieldFromState clears a field from state if write-only version is set
func clearWriteOnlyFieldFromState(version types.Int64, stateValue *types.String) {
	if !version.IsNull() {
		*stateValue = types.StringNull()
	}
}

// handleWriteOnlyFieldUpdates processes all write-only field updates,
func handleWriteOnlyFieldUpdates(
	config *OnePasswordItemResourceModel,
	state *OnePasswordItemResourceModel,
	plan *OnePasswordItemResourceModel,
	refreshItem func() (*model.Item, error),
) error {
	// Check if any write-only field needs the current item
	passwordNeedsItem := shouldFetchCurrentItem(config.PasswordWOVersion, state.PasswordWOVersion)
	noteValueNeedsItem := shouldFetchCurrentItem(config.NoteValueWOVersion, state.NoteValueWOVersion)

	// Fetch item once if needed
	var currentItem *model.Item
	if passwordNeedsItem || noteValueNeedsItem {
		var err error
		currentItem, err = refreshItem()
		if err != nil {
			return err
		}
	}

	// Handle password_wo
	if !config.PasswordWOVersion.IsNull() {
		configVer := config.PasswordWOVersion.ValueInt64()
		stateVer := int64(0)
		if !state.PasswordWOVersion.IsNull() {
			stateVer = state.PasswordWOVersion.ValueInt64()
		}

		if configVer > stateVer {
			plan.Password = config.PasswordWO
		} else {
			fieldFound := false
			for _, f := range currentItem.Fields {
				if f.Purpose == model.FieldPurposePassword {
					plan.Password = types.StringValue(f.Value)
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				plan.Password = types.StringNull()
			}
		}
	}

	// Handle note_value_wo
	if !config.NoteValueWOVersion.IsNull() {
		configVer := config.NoteValueWOVersion.ValueInt64()
		stateVer := int64(0)
		if !state.NoteValueWOVersion.IsNull() {
			stateVer = state.NoteValueWOVersion.ValueInt64()
		}

		if configVer > stateVer {
			plan.NoteValue = config.NoteValueWO
		} else {
			fieldFound := false
			for _, f := range currentItem.Fields {
				if f.Purpose == model.FieldPurposeNotes {
					plan.NoteValue = types.StringValue(f.Value)
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				plan.NoteValue = types.StringNull()
			}
		}
	}

	return nil
}

// shouldFetchCurrentItem returns true if the write-only field version hasn't increased
func shouldFetchCurrentItem(configVersion, stateVersion types.Int64) bool {
	if configVersion.IsNull() {
		return false
	}
	configVer := configVersion.ValueInt64()
	stateVer := int64(0)
	if !stateVersion.IsNull() {
		stateVer = stateVersion.ValueInt64()
	}
	return configVer <= stateVer
}

// writeOnlySectionField is a section field the plan marks write-only, paired with its value_wo
// from configuration. Sections and fields are matched by label, which the map and list forms
// both carry and which stays put when a list is reordered.
type writeOnlySectionField struct {
	sectionLabel string
	fieldLabel   string
	fieldID      types.String
	version      types.Int64
	value        types.String
	setPlanValue func(types.String)
}

// writeOnlySectionFields lists the plan's fields that have a value_wo_version.
//
// The version is taken from the plan rather than the configuration so that ignore_changes is
// respected. Under ignore_changes the plan keeps the prior state's fields, and a field that
// was not write-only before stays that way, keeping its value, instead of being overwritten by
// a value_wo that Terraform was told to ignore.
//
// value_wo itself comes from the configuration, because write-only values are never in the
// plan. The configuration has the same sections and fields as the plan, by key for
// section_map and by position for section, except where ignore_changes has kept fields the
// configuration no longer has; those get a null value_wo.
func writeOnlySectionFields(config, plan *OnePasswordItemResourceModel) []writeOnlySectionField {
	var fields []writeOnlySectionField

	for sectionLabel, planSection := range plan.SectionMap {
		for fieldLabel, planField := range planSection.FieldMap {
			if planField.ValueWOVersion.IsNull() {
				continue
			}
			value := types.StringNull()
			if configSection, ok := config.SectionMap[sectionLabel]; ok {
				if configField, ok := configSection.FieldMap[fieldLabel]; ok {
					value = configField.ValueWO
				}
			}
			fields = append(fields, writeOnlySectionField{
				sectionLabel: sectionLabel,
				fieldLabel:   fieldLabel,
				fieldID:      planField.ID,
				version:      planField.ValueWOVersion,
				value:        value,
				setPlanValue: func(v types.String) {
					planField.Value = v
					planSection.FieldMap[fieldLabel] = planField
				},
			})
		}
	}

	for i, planSection := range plan.SectionList {
		planFields := planSection.FieldList
		for j, planField := range planFields {
			if planField.ValueWOVersion.IsNull() {
				continue
			}
			value := types.StringNull()
			if i < len(config.SectionList) && j < len(config.SectionList[i].FieldList) {
				value = config.SectionList[i].FieldList[j].ValueWO
			}
			fields = append(fields, writeOnlySectionField{
				sectionLabel: planSection.Label.ValueString(),
				fieldLabel:   planField.Label.ValueString(),
				fieldID:      planField.ID,
				version:      planField.ValueWOVersion,
				value:        value,
				setPlanValue: func(v types.String) {
					planFields[j].Value = v
				},
			})
		}
	}

	return fields
}

// stateFieldVersion returns the value_wo_version a section field had in state, or null when the
// field was not in state or did not use value_wo.
func stateFieldVersion(state *OnePasswordItemResourceModel, sectionLabel, fieldLabel string) types.Int64 {
	if state == nil {
		return types.Int64Null()
	}
	if section, ok := state.SectionMap[sectionLabel]; ok {
		if field, ok := section.FieldMap[fieldLabel]; ok {
			return field.ValueWOVersion
		}
	}
	for _, section := range state.SectionList {
		if section.Label.ValueString() != sectionLabel {
			continue
		}
		for _, field := range section.FieldList {
			if field.Label.ValueString() == fieldLabel {
				return field.ValueWOVersion
			}
		}
	}
	return types.Int64Null()
}

// itemSectionFieldValue finds a section field's current value in 1Password, by field ID when the
// field has one and by section and field label otherwise.
func itemSectionFieldValue(item *model.Item, fieldID types.String, sectionLabel, fieldLabel string) (string, bool) {
	if id := fieldID.ValueString(); id != "" {
		for _, f := range item.Fields {
			if f.SectionID != "" && f.ID == id {
				return f.Value, true
			}
		}
	}
	for _, f := range item.Fields {
		if f.SectionID == "" || f.Label != fieldLabel {
			continue
		}
		for _, s := range item.Sections {
			if s.ID == f.SectionID && s.Label == sectionLabel {
				return f.Value, true
			}
		}
	}
	return "", false
}

// applySectionFieldWriteOnlyValues sets the plan value of every section field that uses value_wo,
// so the item sent to 1Password carries it. A field takes its value_wo when it is new or its
// value_wo_version has increased; otherwise it keeps the value 1Password already holds, which is
// read only if some field needs it. state is nil on create.
func applySectionFieldWriteOnlyValues(
	config *OnePasswordItemResourceModel,
	state *OnePasswordItemResourceModel,
	plan *OnePasswordItemResourceModel,
	refreshItem func() (*model.Item, error),
) error {
	var currentItem *model.Item

	for _, f := range writeOnlySectionFields(config, plan) {
		stateVersion := stateFieldVersion(state, f.sectionLabel, f.fieldLabel)
		due := stateVersion.IsNull() || f.version.ValueInt64() > stateVersion.ValueInt64()
		if due && !f.value.IsNull() && !f.value.IsUnknown() {
			f.setPlanValue(f.value)
			continue
		}

		if currentItem == nil {
			var err error
			currentItem, err = refreshItem()
			if err != nil {
				return err
			}
		}
		if value, ok := itemSectionFieldValue(currentItem, f.fieldID, f.sectionLabel, f.fieldLabel); ok {
			f.setPlanValue(types.StringValue(value))
		} else {
			f.setPlanValue(types.StringNull())
		}
	}

	return nil
}

// clearSectionFieldWriteOnlyValues removes from state the value of every section field that has a
// value_wo_version, so a write-only value never reaches state after it is written or read.
func clearSectionFieldWriteOnlyValues(m *OnePasswordItemResourceModel) {
	for _, section := range m.SectionMap {
		for label, field := range section.FieldMap {
			if !field.ValueWOVersion.IsNull() {
				field.Value = types.StringNull()
				section.FieldMap[label] = field
			}
		}
	}
	for i := range m.SectionList {
		for j := range m.SectionList[i].FieldList {
			if !m.SectionList[i].FieldList[j].ValueWOVersion.IsNull() {
				m.SectionList[i].FieldList[j].Value = types.StringNull()
			}
		}
	}
}
