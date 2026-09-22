package provider

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
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
	if !originalValue.IsNull() && !originalValue.IsUnknown() && originalValue.ValueString() == "" && value == "" {
		return types.StringValue("")
	}
	// Original behavior is to convert empty to null
	return setStringValue(value)
}

// setTimeValue formats a timestamp as an RFC 3339 string in UTC. A zero time is
// treated as "not reported by the server" and becomes null.
func setTimeValue(value time.Time) basetypes.StringValue {
	if value.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(value.UTC().Format(time.RFC3339))
}
