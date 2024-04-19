// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/1Password/terraform-provider-onepassword/internal/onepassword"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure OnePasswordProvider satisfies various provider interfaces.
var _ provider.Provider = &OnePasswordProvider{}
var _ provider.ProviderWithFunctions = &OnePasswordProvider{}

// OnePasswordProvider defines the provider implementation.
type OnePasswordProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OnePasswordProviderModel describes the provider data model.
type OnePasswordProviderModel struct {
	ConnectHost         types.String `tfsdk:"url"`
	ConnectToken        types.String `tfsdk:"token"`
	ServiceAccountToken types.String `tfsdk:"service_account_token"`
	Account             types.String `tfsdk:"account"`
	OpCLIPath           types.String `tfsdk:"op_cli_path"`
}

func (p *OnePasswordProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "onepassword"
	resp.Version = p.version
}

func (p *OnePasswordProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The HTTP(S) URL where your 1Password Connect server can be found. Can also be sourced `OP_CONNECT_HOST` environment variable. Provider will use 1Password Connect server if set.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "A valid token for your 1Password Connect server. Can also be sourced from `OP_CONNECT_TOKEN` environment variable. Provider will use 1Password Connect server if set.",
				Optional:            true,
				Sensitive:           true,
				//Validators: []validator.String{
				//	stringvalidator.AlsoRequires(path.Expressions{
				//		path.MatchRoot("url"),
				//	}...),
				//},
			},
			"service_account_token": schema.StringAttribute{
				MarkdownDescription: "A valid 1Password service account token. Can also be sourced from `OP_SERVICE_ACCOUNT_TOKEN` environment variable. Provider will use the 1Password CLI if set.",
				Optional:            true,
				Sensitive:           true,
				//Validators: []validator.String{
				//	stringvalidator.AtLeastOneOf(path.Expressions{
				//		path.MatchRoot("token"),
				//		path.MatchRoot("account"),
				//	}...),
				//},
			},
			"account": schema.StringAttribute{
				Description: "A valid account's sign-in address or ID to use biometrics unlock. Can also be sourced from `OP_ACCOUNT` environment variable. Provider will use the 1Password CLI if set.",
				Optional:    true,
				//Validators: []validator.String{
				//	stringvalidator.ConflictsWith(path.Expressions{
				//		path.MatchRoot("service_account_token"),
				//	}...),
				//},
			},
			"op_cli_path": schema.StringAttribute{
				Description: "The path to the 1Password CLI binary. Can also be sourced from `OP_CLI_PATH` environment variable. Defaults to `op`.",
				Optional:    true,
			},
		},
	}
}

func (p *OnePasswordProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config OnePasswordProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	connectHost := os.Getenv("OP_CONNECT_HOST")
	connectToken := os.Getenv("OP_CONNECT_TOKEN")
	serviceAccountToken := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	account := os.Getenv("OP_ACCOUNT")
	opCLIPath := os.Getenv("OP_CLI_PATH")
	if opCLIPath == "" {
		opCLIPath = "op"
	}

	// Configuration values are now available.
	if !config.ConnectHost.IsNull() {
		connectHost = config.ConnectHost.ValueString()
	}
	if !config.ConnectToken.IsNull() {
		connectToken = config.ConnectToken.ValueString()
	}
	if !config.ServiceAccountToken.IsNull() {
		serviceAccountToken = config.ServiceAccountToken.ValueString()
	}
	if !config.Account.IsNull() {
		account = config.Account.ValueString()
	}
	if !config.OpCLIPath.IsNull() {
		opCLIPath = config.OpCLIPath.ValueString()
	}

	// This is not handled by setting Required to true because Terraform does not handle
	// multiple required attributes well. If only one is set in the provider configuration,
	// the other one is prompted for, but Terraform then forgets the value for the one that
	// is defined in the code. This confusing user-experience can be avoided by handling the
	// requirement of one of the attributes manually.
	if serviceAccountToken != "" || account != "" {
		if connectToken != "" || connectHost != "" {
			resp.Diagnostics.AddError("Config conflict", "Either Connect credentials (\"token\" and \"url\") or 1Password CLI (\"service_account_token\" or \"account\") credentials can be set. Both are set. Please unset one of them.")
		}
		if opCLIPath == "" {
			resp.Diagnostics.AddAttributeError(path.Root("op_cli_path"), "CLI path missing", "Path to op CLI binary is not set. Either leave empty, provide the \"op_cli_path\" field in the provider configuration, or set the OP_CLI_PATH environment variable.")
		}
		if serviceAccountToken != "" && account != "" {
			resp.Diagnostics.AddError("Config conflict", "\"service_account_token\" and \"account\" are set. Please unset one of them to use the provider with 1Password CLI.")
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	client, err := onepassword.NewClient(onepassword.ClientConfig{
		ConnectHost:         connectHost,
		ConnectToken:        connectToken,
		ServiceAccountToken: serviceAccountToken,
		Account:             account,
		OpCLIPath:           opCLIPath,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client init failure", fmt.Sprintf("Client failed to initialize, got error: %s", err))
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OnePasswordProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOnePasswordItemResource,
	}
}

func (p *OnePasswordProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOnePasswordItemDataSource,
		NewOnePasswordVaultDataSource,
	}
}

func (p *OnePasswordProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OnePasswordProvider{
			version: version,
		}
	}
}
