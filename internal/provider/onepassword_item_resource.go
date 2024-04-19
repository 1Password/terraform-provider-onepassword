// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OnePasswordItemResource{}
var _ resource.ResourceWithImportState = &OnePasswordItemResource{}

func NewOnePasswordItemResource() resource.Resource {
	return &OnePasswordItemResource{}
}

// OnePasswordItemResource defines the resource implementation.
type OnePasswordItemResource struct {
	client *http.Client
}

// ExampleResourceModel describes the resource data model.
type ExampleResourceModel struct {
	ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	Defaulted             types.String `tfsdk:"defaulted"`
	Id                    types.String `tfsdk:"id"`
}

func (r *OnePasswordItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *OnePasswordItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	passwordRecipeSchema := schema.ListNestedBlock{
		Description: passwordRecipeDescription,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"length": schema.Int64Attribute{
					Description: passwordLengthDescription,
					Optional:    true,
					Default:     int64default.StaticInt64(32),
					Validators: []validator.Int64{
						int64validator.Between(1, 64),
					},
				},
				"letters": schema.BoolAttribute{
					Description: passwordLettersDescription,
					Optional:    true,
					Default:     booldefault.StaticBool(true),
				},
				"digits": schema.BoolAttribute{
					Description: passwordDigitsDescription,
					Optional:    true,
					Default:     booldefault.StaticBool(true),
				},
				"symbols": schema.BoolAttribute{
					Description: passwordSymbolsDescription,
					Optional:    true,
					Default:     booldefault.StaticBool(true),
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A 1Password Item",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: terraformItemIDDescription,
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: itemUUIDDescription,
				Computed:            true,
			},
			"vault": schema.StringAttribute{
				MarkdownDescription: vaultUUIDDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"category": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(enumDescription, categoryDescription, categories),
				Optional:            true,
				Default:             stringdefault.StaticString("login"),
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(categories...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: itemTitleDescription,
				Optional:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: urlDescription,
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: dbHostnameDescription,
				Optional:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: dbDatabaseDescription,
				Optional:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: dbPortDescription,
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(enumDescription, dbTypeDescription, dbTypes),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(dbTypes...),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: tagsDescription,
				ElementType:         types.StringType,
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: usernameDescription,
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: passwordDescription,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			// TODO: See if we want to have this attribute in the resource schema.
			//       It exists in the data source schema.
			//"note_value": schema.StringAttribute{
			//	MarkdownDescription: noteValueDescription,
			//	Optional:            true,
			//	Computed:            true,
			//	Sensitive:           true,
			//},
		},
		Blocks: map[string]schema.Block{
			"section": schema.ListNestedBlock{
				MarkdownDescription: sectionsDescription,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: sectionIDDescription,
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: sectionLabelDescription,
							Required:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"field": schema.ListNestedBlock{
							MarkdownDescription: sectionFieldsDescription,
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fieldIDDescription,
										Optional:            true,
										Computed:            true,
									},
									"label": schema.StringAttribute{
										MarkdownDescription: fieldLabelDescription,
										Required:            true,
									},
									"purpose": schema.StringAttribute{
										MarkdownDescription: fmt.Sprintf(enumDescription, fieldPurposeDescription, fieldPurposes),
										Optional:            true,
										Validators: []validator.String{
											stringvalidator.OneOfCaseInsensitive(fieldPurposes...),
										},
									},
									"type": schema.StringAttribute{
										MarkdownDescription: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
										Optional:            true,
										Default:             stringdefault.StaticString("STRING"),
										Validators: []validator.String{
											stringvalidator.OneOfCaseInsensitive(fieldTypes...),
										},
									},
									"value": schema.StringAttribute{
										MarkdownDescription: fieldValueDescription,
										Optional:            true,
										Computed:            true,
										Sensitive:           true,
									},
								},
								Blocks: map[string]schema.Block{
									"password_recipe": passwordRecipeSchema,
								},
							},
						},
					},
				},
			},
			"password_recipe": passwordRecipeSchema,
		},
	}
}

func (r *OnePasswordItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OnePasswordItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExampleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnePasswordItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExampleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnePasswordItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExampleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnePasswordItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExampleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *OnePasswordItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
