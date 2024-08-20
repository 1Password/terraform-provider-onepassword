package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OnePasswordVaultDataSource{}

func NewOnePasswordVaultDataSource() datasource.DataSource {
	return &OnePasswordVaultDataSource{}
}

// OnePasswordVaultDataSource defines the data source implementation.
type OnePasswordVaultDataSource struct {
	client onepassword.Client
}

// OnePasswordVaultDataSourceModel describes the data source data model.
type OnePasswordVaultDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *OnePasswordVaultDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (d *OnePasswordVaultDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Use this data source to get details of a vault by either its name or uuid.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Terraform resource identifier for this item in the format `vaults/<vault_id>`",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the vault to retrieve. This field will be populated with the UUID of the vault if the vault it looked up by its name.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
						path.MatchRoot("uuid"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the vault to retrieve. This field will be populated with the name of the vault if the vault it looked up by its UUID.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the vault.",
				Computed:            true,
			},
		},
	}
}

func (d *OnePasswordVaultDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(onepassword.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected onepassword.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OnePasswordVaultDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnePasswordVaultDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var vault *op.Vault
	if data.UUID.ValueString() != "" {
		vaultByUUID, err := d.client.GetVault(ctx, data.UUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vault, got error: %s", err))
			return
		}
		vault = vaultByUUID
	} else {
		vaultsByName, err := d.client.GetVaultsByTitle(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vault, got error: %s", err))
			return
		}
		if len(vaultsByName) == 0 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("No vault found with name '%s'", data.Name))
			return
		} else if len(vaultsByName) > 1 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Multiple vaults found with name '%s'", data.Name))
			return
		}
		fullVault, err := d.client.GetVault(ctx, vaultsByName[0].ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vault, got error: %s", err))
			return
		}
		vault = fullVault
	}

	data = OnePasswordVaultDataSourceModel{
		ID:          types.StringValue(vaultTerraformID(vault)),
		UUID:        types.StringValue(vault.ID),
		Name:        types.StringValue(vault.Name),
		Description: types.StringValue(vault.Description),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
