package onepassword

import (
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider The 1Password Connect terraform provider
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The HTTP(S) Url where your 1Password Connect API can be found",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_TOKEN", nil),
				Description: "A valid token for your 1Password Connect API",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"onepassword_item": dataSourceOnepasswordItem(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"onepassword_item": resourceOnepasswordItem(),
		},
	}
	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return connect.NewClient(d.Get("url").(string), d.Get("token").(string)), nil
	}
	return provider
}
