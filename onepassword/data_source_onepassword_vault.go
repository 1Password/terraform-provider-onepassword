package onepassword

import (
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var oneOfUUIDName = []string{"uuid", "name"}

func dataSourceOnepasswordVault() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get details of a vault by either its name or uuid.",
		Read:        dataSourceOnepasswordVaultRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The Terraform resource identifier for this item in the format `vaults/<vault_id>`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"uuid": {
				Description:  "The UUID of the vault to retrieve. This field will be populated with the UUID of the vault if the vault it looked up by its name.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: oneOfUUIDName,
			},
			"name": {
				Description:  "The name of the vault to retrieve. This field will be populated with the name of the vault if the vault it looked up by its UUID.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: oneOfUUIDName,
			},
			"description": {
				Description: "The description of the vault.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceOnepasswordVaultRead(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)

	title := data.Get("name").(string)
	vaultUUID := data.Get("uuid").(string)

	var vault *onepassword.Vault
	if vaultUUID != "" {
		vaultByUUID, err := client.GetVault(vaultUUID)
		if err != nil {
			return err
		}
		vault = vaultByUUID
	} else {
		vaultsByName, err := client.GetVaultsByTitle(title)
		if err != nil {
			return err
		}
		if len(vaultsByName) == 0 {
			return fmt.Errorf("no vault found with name '%s'", title)
		} else if len(vaultsByName) > 1 {
			return fmt.Errorf("multiple vaults found with name '%s'", title)
		}
		vault = &vaultsByName[0]
	}

	data.SetId(vaultTerraformID(vault))
	data.Set("uuid", vault.ID)
	data.Set("name", vault.Name)
	data.Set("description", vault.Description)

	return nil
}

func vaultTerraformID(vault *onepassword.Vault) string {
	return fmt.Sprintf("vaults/%s", vault.ID)
}
