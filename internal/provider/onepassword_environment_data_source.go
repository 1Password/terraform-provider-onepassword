package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OnePasswordEnvironmentDataSource{}

func NewOnePasswordEnvironmentDataSource() datasource.DataSource {
	return &OnePasswordEnvironmentDataSource{}
}

// OnePasswordEnvironmentDataSource defines the data source implementation.
type OnePasswordEnvironmentDataSource struct {
	client onepassword.Client
}

// OnePasswordEnvironmentDataSourceModel describes the data source data model.
type OnePasswordEnvironmentDataSourceModel struct {
	ID            types.String       `tfsdk:"id"`
	EnvironmentID types.String       `tfsdk:"environment_id"`
	Variables     types.Map          `tfsdk:"variables"`
	Metadata      []envVariableModel `tfsdk:"metadata"`
}

type envVariableModel struct {
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Masked types.Bool   `tfsdk:"masked"`
}

func (d *OnePasswordEnvironmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *OnePasswordEnvironmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read environment variables from a [1Password Environment](https://developer.1password.com/docs/environments/). " +
			"1Password Environments allow you to organize and manage project secrets as environment variables. " +
			"This data source is only supported when using **service account** or **desktop app** authentication; it is not available with 1Password Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Terraform resource identifier for this environment in the format `environments/<environment_id>`.",
				Computed:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the 1Password Environment. You can find this in the 1Password desktop app under Developer > View Environments > Manage environment > Copy environment ID.",
				Required:            true,
			},
			"variables": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of environment variable names to their values. Use this for passing secrets into Terraform resources or for use in `environment` blocks.",
				Computed:            true,
				Sensitive:           true,
			},
		},
		Blocks: map[string]schema.Block{
			"metadata": schema.ListNestedBlock{
				MarkdownDescription: "Metadata for each environment variable (name, value, and masked flag). Use this when you need the full structure; use `variables` for a simple name-to-value map.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The environment variable name.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The environment variable value.",
							Computed:            true,
							Sensitive:           true,
						},
						"masked": schema.BoolAttribute{
							MarkdownDescription: "Whether the value is hidden by default in the 1Password app.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *OnePasswordEnvironmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OnePasswordEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnePasswordEnvironmentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := data.EnvironmentID.ValueString()
	if environmentID == "" {
		resp.Diagnostics.AddError("Missing environment_id", "The environment_id attribute is required.")
		return
	}

	variables, err := d.client.GetEnvironmentVariables(ctx, environmentID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read 1Password Environment variables, got error: %s", err))
		return
	}

	variablesMap := make(map[string]string)
	metadataList := make([]envVariableModel, 0, len(variables))

	for _, v := range variables {
		variablesMap[v.Name] = v.Value
		metadataList = append(metadataList, envVariableModel{
			Name:   types.StringValue(v.Name),
			Value:  types.StringValue(v.Value),
			Masked: types.BoolValue(v.Masked),
		})
	}

	data.ID = types.StringValue(environmentTerraformID(environmentID))
	data.EnvironmentID = types.StringValue(environmentID)

	variablesMapVal, diags := types.MapValueFrom(ctx, types.StringType, variablesMap)
	resp.Diagnostics.Append(diags...)
	data.Variables = variablesMapVal
	data.Metadata = metadataList

	tflog.Trace(ctx, "read 1Password Environment data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func environmentTerraformID(environmentID string) string {
	return "environments/" + environmentID
}
