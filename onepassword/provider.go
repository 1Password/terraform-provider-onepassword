package onepassword

import (
	"context"
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/1Password/terraform-provider-onepassword/version"
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
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_HOST", nil),
				Description: "The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the the OP_CONNECT_HOST environment variable if this attribute is not set.",
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
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		url := d.Get("url").(string)
		token := d.Get("token").(string)

		// This is not handled by setting Required to true because Terraform does not handle
		// multiple required attributes well. If only one is set in the provider configuration,
		// the other one is prompted for, but Terraform then forgets the value for the one that
		// is defined in the code. This confusing user-experience can be avoided by handling the
		// requirement of one of the attributes manually.
		if url == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "URL for Connect API is not set",
				Detail:   "Either provide the \"url\" field in the provider configuration or set the OP_CONNECT_HOST environment variable",
			})
			return nil, diags
		}

		return connect.NewClientWithUserAgent(url, token, providerUserAgent), nil
	}
	return provider
}
