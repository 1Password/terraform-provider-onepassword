package onepassword

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/terraform-provider-onepassword/opcli"
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
			"account": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_ACCOUNT", nil),
				Description: "The account to execute the command by account shorthand, sign-in address, account UUID, or user UUID.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_PASSWORD", nil),
				Description: "The password to interact with the CLI",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_HOST", nil),
				Description: "The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the the OP_CONNECT_HOST environment variable if this attribute is not set.",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
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
		var op bool
		url := d.Get("url").(string)
		token := d.Get("token").(string)
		account := d.Get("account").(string)
		password := d.Get("password").(string)

		if _, err := exec.LookPath("op"); err == nil {
			op = true
		}

		if url != "" || token != "" {
			if url == "" {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  "URL for Connect API is not set",
					Detail:   "Either provide the \"url\" field in the provider configuration or set the OP_CONNECT_HOST environment variable",
				}}
			}
			if token == "" {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  "TOKEN for Connect API is not set",
					Detail:   "Either provide the \"token\" field in the provider configuration or set the OP_CONNECT_TOKEN environment variable",
				}}
			}
			return connect.NewClientWithUserAgent(url, token, providerUserAgent), nil
		} else if account == "" {
			return nil, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "ACCOUNT is not set",
				Detail:   "Either provide the \"account\" field in the provider configuration or set the OP_ACCOUNT environment variable",
			}}
		} else if !op {
			return nil, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "op executable not found",
				Detail:   "Please ensure you have the 1password-cli >= 2.0.0 installed in your $PATH.",
			}}
		} else if password == "" {
			return nil, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Password is not set",
				Detail:   "Provide the OP_PASSWORD environment variable.",
			}}
		} else {
			provider, err := opcli.NewCLIClient(account, password)
			if err != nil {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  "Could not initialize CLI provider",
					Detail:   err.Error(),
				}}
			}

			return provider, nil
		}
	}
	return provider
}
