package onepassword

import (
	"context"
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/onepassword/cli"
	"github.com/1Password/terraform-provider-onepassword/onepassword/connectctx"
	"github.com/1Password/terraform-provider-onepassword/version"
	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	terraformProviderUserAgent = "terraform-provider-connect/%s"
	minimumOpCliVersion        = "2.23.0" // introduction of stdin json support for `op item update`
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

// Provider The 1Password Terraform provider
func Provider() *schema.Provider {
	providerUserAgent := fmt.Sprintf(terraformProviderUserAgent, version.ProviderVersion)
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_HOST", nil),
				Description: "The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the OP_CONNECT_HOST environment variable if this attribute is not set. Must be set to use with 1Password Connect server.",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CONNECT_TOKEN", nil),
				Description: "A valid token for your 1Password Connect API. Can also be sourced from OP_CONNECT_TOKEN. Must be set to use with 1Password Connect server.",
			},
			"service_account_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_SERVICE_ACCOUNT_TOKEN", nil),
				Description: "A valid token for your 1Password Service Account. Can also be sourced from OP_SERVICE_ACCOUNT_TOKEN. Must be set to use with 1Password service account.",
			},
			"account": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_ACCOUNT", nil),
				Description: "A valid account's sign-in address or ID to use biometrics unlock. Can also be sourced from OP_ACCOUNT. Must be set to use with biometric unlock.",
			},
			"op_cli_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OP_CLI_PATH", "op"),
				Description: "The path to the 1Password CLI binary. Can also be sourced from OP_CLI_PATH. Defaults to `op`.",
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
			account             = d.Get("account").(string)
			opCliPath           = d.Get("op_cli_path").(string)
		)

		// This is not handled by setting Required to true because Terraform does not handle
		// multiple required attributes well. If only one is set in the provider configuration,
		// the other one is prompted for, but Terraform then forgets the value for the one that
		// is defined in the code. This confusing user-experience can be avoided by handling the
		// requirement of one of the attributes manually.
		if serviceAccountToken != "" {
			if token != "" || url != "" {
				return nil, diag.Errorf("Either Connect credentials (\"token\" and \"url\") or Service Account (\"service_account_token\") credentials can be set. Both are set. Please unset one of them.")
			}
			if opCliPath == "" {
				return nil, diag.Errorf("Path to op CLI binary is not set. Either leave empty, provide the \"op_cli_path\" field in the provider configuration, or set the OP_CLI_PATH environment variable.")
			}

			op := cli.New(serviceAccountToken, opCliPath, "")

			cliVersion, err := op.GetVersion(ctx)
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("failed to get version of op CLI: %w", err))
			}
			if cliVersion.LessThan(semver.MustParse(minimumOpCliVersion)) {
				return nil, diag.Errorf("Current 1Password CLI version is \"%s\". Please upgrade to at least \"%s\".", cliVersion, minimumOpCliVersion)
			}

			return (Client)(op), nil
		} else if account != "" {
			op := cli.New("", opCliPath, account)
			cliVersion, err := op.GetVersion(ctx)
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("failed to get version of op CLI: %w", err))
			}
			if cliVersion.LessThan(semver.MustParse(minimumOpCliVersion)) {
				return nil, diag.Errorf("Current 1Password CLI version is \"%s\". Please upgrade to at least \"%s\".", cliVersion, minimumOpCliVersion)
			}

			return (Client)(op), nil
		} else if token != "" && url != "" {
			return connectctx.Wrap(connect.NewClientWithUserAgent(url, token, providerUserAgent)), nil
		} else {
			return nil, diag.Errorf("Invalid provider configuration. Either Connect credentials (\"token\" and \"url\") or Service Account (\"service_account_token\" or \"account\") credentials should be set.")
		}
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
