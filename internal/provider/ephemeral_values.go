package provider

import (
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
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

// func (r *OnePasswordItemResource) handleWriteOnlyFieldUpdate(
// 	ctx context.Context,
// 	configVersion types.Int64,
// 	stateVersion types.Int64,
// 	woValue types.String,
// 	planValue *types.String,
// 	planID types.String,
// 	fieldPurpose model.ItemFieldPurpose,
// 	fieldName string,
// ) error {
// 	if configVersion.IsNull() {
// 		return nil
// 	}

// 	configVer := configVersion.ValueInt64()
// 	stateVer := int64(0)
// 	if !stateVersion.IsNull() {
// 		stateVer = stateVersion.ValueInt64()
// 	}

// 	if configVer > stateVer {
// 		// Version increased - use new write-only value
// 		*planValue = woValue
// 	} else {
// 		// Version unchanged or decreased - preserve existing value by reading current item
// 		vaultUUID, itemUUID := vaultAndItemUUID(planID.ValueString())
// 		currentItem, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
// 		if err != nil {
// 			return fmt.Errorf("Could not read item '%s' from vault '%s' to preserve %s, got error: %s", itemUUID, vaultUUID, fieldName, err)
// 		}
// 		// Extract field from current item
// 		fieldFound := false
// 		for _, f := range currentItem.Fields {
// 			if f.Purpose == fieldPurpose {
// 				*planValue = types.StringValue(f.Value)
// 				fieldFound = true
// 				break
// 			}
// 		}
// 		// Field not found (user removed it in 1Password), sync to that state
// 		if !fieldFound {
// 			*planValue = types.StringNull()
// 		}
// 	}
// 	return nil
// }

// handleWriteOnlyFieldUpdates processes all write-only field updates,
func handleWriteOnlyFieldUpdates(
	config *OnePasswordItemResourceModel,
	state *OnePasswordItemResourceModel,
	plan *OnePasswordItemResourceModel,
	refreshItem func() (*model.Item, error),
) error {
	// Helper to check if a field needs the current item
	needsItem := func(configVersion, stateVersion types.Int64) bool {
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

	// Check if any write-only field needs the current item
	passwordNeedsItem := needsItem(config.PasswordWOVersion, state.PasswordWOVersion)
	noteValueNeedsItem := needsItem(config.NoteValueWOVersion, state.NoteValueWOVersion)

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
