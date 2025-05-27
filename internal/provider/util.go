package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/model"
)

func vaultTerraformID(vault *model.Vault) string {
	return fmt.Sprintf("vaults/%s", vault.ID)
}

func itemTerraformID(item *model.Item) string {
	return fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID)
}

func setStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
