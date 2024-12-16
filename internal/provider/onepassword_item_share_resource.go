package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OnePasswordItemShareResource{}
var _ resource.ResourceWithImportState = &OnePasswordItemShareResource{}

func NewOnePasswordItemShareResource() resource.Resource {
	return &OnePasswordItemShareResource{}
}

// OnePasswordItemShareResource defines the resource implementation.
type OnePasswordItemShareResource struct {
	client onepassword.Client
}

// OnePasswordItemShareResourceModel describes the resource data model.
type OnePasswordItemShareResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Item      types.String `tfsdk:"item"`
	Vault     types.String `tfsdk:"vault"`
	Emails    types.String `tfsdk:"emails"`
	ExpiresIn types.String `tfsdk:"expires_in"`
	ViewOnce  types.Bool   `tfsdk:"view_once"`
	ShareURL  types.String `tfsdk:"share_url"`
}

func (r *OnePasswordItemShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_share"
}

func (r *OnePasswordItemShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
		A 1Password Item Share URL.

> **Note:** Sharing at item is only supported by the 1Password CLI, and therefore, only by service and user accounts.
Attempting to create this resource when using 1Password Connect Server will result in an error.`,

		Attributes: map[string]schema.Attribute{
			// "id": schema.StringAttribute{
			// 	MarkdownDescription: terraformItemIDDescription,
			// 	Computed:            true,
			// },
			"id": schema.StringAttribute{
				MarkdownDescription: terraformItemIDDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"item": schema.StringAttribute{
				MarkdownDescription: itemUUIDDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vault": schema.StringAttribute{
				MarkdownDescription: vaultUUIDDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"emails": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of emails to allow to access the item",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,},\s*)*([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$`),
						"must be a comma-separated list of valid email addresses",
					),
				},
			},
			"expires_in": schema.StringAttribute{
				MarkdownDescription: "The time until the share expires. Must be a number followed by a time unit (s, m, h, d, w). Not valid when `view_once` is set to `true`",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+[smhdw]$`),
						"must be a number followed by a time unit (s, m, h, d, w)",
					),
				},
			},
			"view_once": schema.BoolAttribute{
				MarkdownDescription: "Whether the share should be viewable only once. Not valid when `expires_in` is set",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"share_url": schema.StringAttribute{
				MarkdownDescription: "The item share URL",
				Computed:            true,
			},
		},
	}
}

func (r *OnePasswordItemShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(onepassword.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected onepassword.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OnePasswordItemShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OnePasswordItemShareResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	shareUrl, err := r.client.ShareItem(ctx, data.Item.ValueString(), data.Vault.ValueString(), data.Emails.ValueString(), data.ExpiresIn.ValueString(), data.ViewOnce.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item Share create error", fmt.Sprintf("Error creating 1Password item share, got error %s", err))
		return
	}

	data.ShareURL = setStringValue(*shareUrl)

	data.ID = setStringValue(fmt.Sprintf("vault/%s/item/%s/share/%s/%s/%t", data.Item.ValueString(), data.Vault.ValueString(), data.Emails.ValueString(), data.ExpiresIn.ValueString(), data.ViewOnce.ValueBool()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// The 1Password Item Share resource does not support reading, updating, deleting, or importing.
//
// These methods are implemented to satisfy the interface requirements, but they are not used, and no warnings will be output.
func (r *OnePasswordItemShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *OnePasswordItemShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *OnePasswordItemShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *OnePasswordItemShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

func (r *OnePasswordItemShareResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data OnePasswordItemShareResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ViewOnce.ValueBool() && !data.ExpiresIn.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("attribute_two"),
			"Conflicting Attributes",
			"The expiration cannot be set when the share is only viewable once",
		)
	}
}
