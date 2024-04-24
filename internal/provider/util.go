package provider

import (
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func vaultTerraformID(vault *op.Vault) string {
	return fmt.Sprintf("vaults/%s", vault.ID)
}

func itemTerraformID(item *op.Item) string {
	return fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID)
}

func setStringValue(value string) basetypes.StringValue {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
