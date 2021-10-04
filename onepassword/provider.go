package onepassword

import (
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/terraform-provider-onepassword/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	terraformProviderUserAgent = "terraform-provider-connect/%s"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

// Provider The 1Password Connect terraform provider
func Provider() *schema.Provider {
	providerUserAgent := fmt.Sprintf(terraformProviderUserAgent, version.ProviderVersion)
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_HOST", nil),
				Description: "The HTTP(S) URL where your 1Password Connect API can be found. Can also be sourced from OP_CONNECT_HOST.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_TOKEN", nil),
				Description: "A valid token for your 1Password Connect API. Can also be sourced from OP_CONNECT_TOKEN.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"onepassword_vault": dataSourceOnepasswordVault(),
			"onepassword_item":  dataSourceOnepasswordItem(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"onepassword_item": resourceOnepasswordItem(),
		},
	}
	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return connect.NewClientWithUserAgent(d.Get("url").(string), d.Get("token").(string), providerUserAgent), nil
	}
	return provider
}
