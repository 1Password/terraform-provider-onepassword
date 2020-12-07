package onepassword

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/1Password/connect-sdk-go/connect"
)

func dataSourceOnepasswordItem() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOnepasswordItemRead,
		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vault": {
				Type:     schema.TypeString,
				Required: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "login",
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"database": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceOnepasswordItemRead(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)

	vaultUUID := data.Get("vault").(string)
	itemUUID := data.Get("uuid").(string)

	data.SetId("")

	item, err := client.GetItem(itemUUID, vaultUUID)

	if err != nil {
		return err
	}

	data.SetId(terraformID(item))
	data.Set("uuid", item.ID)
	data.Set("vault", item.Vault.ID)
	data.Set("title", item.Title)

	for _, u := range item.URLs {
		if u.Primary {
			data.Set("url", u.URL)
		}
	}

	data.Set("tags", item.Tags)
	data.Set("category", strings.ToLower(string(item.Category)))

	for _, f := range item.Fields {
		switch f.Purpose {
		case "USERNAME":
			data.Set("username", f.Value)
		case "PASSWORD":
			data.Set("password", f.Value)
		default:
			data.Set(strings.ToLower(f.Label), f.Value)
		}
	}

	return nil
}
