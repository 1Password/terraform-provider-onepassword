package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func vaultTerraformID(vault *model.Vault) string {
	return fmt.Sprintf("vaults/%s", vault.ID)
}

func itemTerraformID(item *model.Item) string {
	return fmt.Sprintf("vaults/%s/items/%s", item.VaultID, item.ID)
}

func setStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

// setStringValuePreservingEmpty preserves empty strings when they were explicitly set in Terraform
func setStringValuePreservingEmpty(value string, originalValue basetypes.StringValue) basetypes.StringValue {
	// If original was explicitly set to empty string (not null), preserve it
	if !originalValue.IsNull() && originalValue.ValueString() == "" && value == "" {
		return types.StringValue("")
	}
	// Original behavior is to convert empty to null
	return setStringValue(value)
}
