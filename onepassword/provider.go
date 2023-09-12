package onepassword

import (
	"context"
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/onepassword/cli"
	"github.com/1Password/terraform-provider-onepassword/onepassword/connectctx"
	"github.com/1Password/terraform-provider-onepassword/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_HOST", nil),
				Description: "The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the OP_CONNECT_HOST environment variable if this attribute is not set. Can be omitted, if service_account_token is set.",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_TOKEN", nil),
				Description: "A valid token for your 1Password Connect API. Can also be sourced from OP_CONNECT_TOKEN. Either this or `service_account_token` must be set.",
			},
			"service_account_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_SERVICE_ACCOUNT_TOKEN", nil),
				Description: "A valid token for your 1Password Service Account. Can also be sourced from OP_SERVICE_ACCOUNT_TOKEN. Either this or `token` must be set.",
			},
			"op_cli_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CLI_PATH", "op"),
				Description: "The path to the 1Password CLI binary. Can also be sourced from OP_CLI_PATH. Defaults to `op`. Only used when setting a `service_account_token`.",
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
		var (
			url                 = d.Get("url").(string)
			token               = d.Get("token").(string)
			serviceAccountToken = d.Get("service_account_token").(string)
			opCliPath           = d.Get("op_cli_path").(string)
		)

		// This is not handled by setting Required to true because Terraform does not handle
		// multiple required attributes well. If only one is set in the provider configuration,
		// the other one is prompted for, but Terraform then forgets the value for the one that
		// is defined in the code. This confusing user-experience can be avoided by handling the
		// requirement of one of the attributes manually.
		if serviceAccountToken != "" {
			if token != "" || url != "" {
				return nil, diag.Errorf("Either (\"token\" and \"url\") or \"service_account_token\" field can be set. Both are set. Please unset one of them.")
			}
			return (Client)(cli.New(serviceAccountToken, opCliPath)), nil
		}

		if token == "" {
			return nil, diag.Errorf("Token for Connect API is not set. Either provide the \"token\" field in the provider configuration or set the OP_CONNECT_TOKEN environment variable")
		}
		if url == "" {
			return nil, diag.Errorf("URL for Connect API is not set. Either provide the \"url\" field in the provider configuration or set the OP_CONNECT_HOST environment variable")
		}

		return connectctx.Wrap(connect.NewClientWithUserAgent(url, token, providerUserAgent)), nil
	}
	return provider
}

// Client is a subset of connect.Client with context added.
type Client interface {
	GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error)
	GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error)
	GetItem(ctx context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error)
	GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*onepassword.Item, error)
	CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error)
	DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error
}
