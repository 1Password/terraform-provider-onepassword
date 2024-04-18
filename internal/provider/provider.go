// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
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
			},
			"service_account_token": schema.StringAttribute{
				MarkdownDescription: "A valid 1Password service account token. Can also be sourced from `OP_SERVICE_ACCOUNT_TOKEN` environment variable. Provider will use the 1Password CLI if set.",
				Optional:            true,
				Sensitive:           true,
			},
			"account": schema.StringAttribute{
				Description: "A valid account's sign-in address or ID to use biometrics unlock. Can also be sourced from `OP_ACCOUNT` environment variable. Provider will use the 1Password CLI if set.",
				Optional:    true,
			},
			"op_cli_path": schema.StringAttribute{
				Description: "The path to the 1Password CLI binary. Can also be sourced from `OP_CLI_PATH` environment variable. Defaults to `op`.",
				Optional:    true,
			},
		},
	}
}

func (p *OnePasswordProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OnePasswordProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OnePasswordProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *OnePasswordProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
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
