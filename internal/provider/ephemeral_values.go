package provider

import (
	"context"
	"fmt"

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

func (r *OnePasswordItemResource) handleWriteOnlyFieldUpdate(
	ctx context.Context,
	configVersion types.Int64,
	stateVersion types.Int64,
	woValue types.String,
	planValue *types.String,
	planID types.String,
	fieldPurpose model.ItemFieldPurpose,
	fieldName string,
) error {
	if configVersion.IsNull() {
		return nil
	}

	configVer := configVersion.ValueInt64()
	stateVer := int64(0)
	if !stateVersion.IsNull() {
		stateVer = stateVersion.ValueInt64()
	}

	if configVer > stateVer {
		// Version increased - use new write-only value
		*planValue = woValue
	} else {
		// Version unchanged or decreased - preserve existing value by reading current item
		vaultUUID, itemUUID := vaultAndItemUUID(planID.ValueString())
		currentItem, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
		if err != nil {
			return fmt.Errorf("Could not read item '%s' from vault '%s' to preserve %s, got error: %s", itemUUID, vaultUUID, fieldName, err)
		}
		// Extract field from current item
		fieldFound := false
		for _, f := range currentItem.Fields {
			if f.Purpose == fieldPurpose {
				*planValue = types.StringValue(f.Value)
				fieldFound = true
				break
			}
		}
		// Field not found (user removed it in 1Password), sync to that state
		if !fieldFound {
			*planValue = types.StringNull()
		}
	}
	return nil
}
