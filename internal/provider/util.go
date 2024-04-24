package provider

import (
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
)

func vaultTerraformID(vault *op.Vault) string {
	return fmt.Sprintf("vaults/%s", vault.ID)
}

func itemTerraformID(item *op.Item) string {
	return fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID)
}
