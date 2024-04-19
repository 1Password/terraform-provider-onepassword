// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/internal/onepassword/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	client onepassword.Client
}

// OnePasswordItemResourceModel describes the resource data model.
type OnePasswordItemResourceModel struct {
	ID       types.String `tfsdk:"id"`
	UUID     types.String `tfsdk:"uuid"`
	Vault    types.String `tfsdk:"vault"`
	Category types.String `tfsdk:"category"`
	Title    types.String `tfsdk:"title"`
	URL      types.String `tfsdk:"url"`
	Hostname types.String `tfsdk:"hostname"`
	Database types.String `tfsdk:"database"`
	Port     types.String `tfsdk:"port"`
	Type     types.String `tfsdk:"type"`
	Tags     types.List   `tfsdk:"tags"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	//NoteValue types.String `tfsdk:"note_value"`
	Section []OnePasswordItemResourceSectionModel `tfsdk:"section"`
	Recipe  PasswordRecipeModel                   `tfsdk:"recipe"`
}

type PasswordRecipeModel struct {
	Length  types.Int64 `tfsdk:"length"`
	Letters types.Bool  `tfsdk:"letters"`
	Digits  types.Bool  `tfsdk:"digits"`
	Symbols types.Bool  `tfsdk:"symbols"`
}

type OnePasswordItemResourceSectionModel struct {
	ID    types.String                        `tfsdk:"id"`
	Label types.String                        `tfsdk:"label"`
	Field []OnePasswordItemResourceFieldModel `tfsdk:"field"`
}

type OnePasswordItemResourceFieldModel struct {
	OnePasswordItemFieldModel
	Recipe PasswordRecipeModel `tfsdk:"recipe"`
}

func (r *OnePasswordItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *OnePasswordItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	//passwordRecipeBlockSchema := schema.ListNestedBlock{
	//	Description: passwordRecipeDescription,
	//	Validators: []validator.List{
	//		listvalidator.SizeAtMost(1),
	//	},
	//	NestedObject: schema.NestedBlockObject{
	//		Attributes: map[string]schema.Attribute{
	//			"length": schema.Int64Attribute{
	//				Description: passwordLengthDescription,
	//				Optional:    true,
	//				Default:     int64default.StaticInt64(32),
	//				Validators: []validator.Int64{
	//					int64validator.Between(1, 64),
	//				},
	//			},
	//			"letters": schema.BoolAttribute{
	//				Description: passwordLettersDescription,
	//				Optional:    true,
	//				Default:     booldefault.StaticBool(true),
	//			},
	//			"digits": schema.BoolAttribute{
	//				Description: passwordDigitsDescription,
	//				Optional:    true,
	//				Default:     booldefault.StaticBool(true),
	//			},
	//			"symbols": schema.BoolAttribute{
	//				Description: passwordSymbolsDescription,
	//				Optional:    true,
	//				Default:     booldefault.StaticBool(true),
	//			},
	//		},
	//	},
	//}

	passwordRecipeAttributeSchema := schema.SingleNestedAttribute{
		Description: passwordRecipeDescription,
		Optional:    true,
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
			"password_recipe": passwordRecipeAttributeSchema,
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
									"password_recipe": passwordRecipeAttributeSchema,
								},
								//Blocks: map[string]schema.Block{
								//	"password_recipe": passwordRecipeBlockSchema,
								//},
							},
						},
					},
				},
			},
			//"password_recipe": passwordRecipeBlockSchema,
		},
	}
}

func (r *OnePasswordItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OnePasswordItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OnePasswordItemResourceModel

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
	data.ID = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnePasswordItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OnePasswordItemResourceModel

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
	var data OnePasswordItemResourceModel

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
	var data OnePasswordItemResourceModel

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

func vaultAndItemUUID(tfID string) (vaultUUID, itemUUID string) {
	elements := strings.Split(tfID, "/")

	if len(elements) != 4 {
		return "", ""
	}

	return elements[1], elements[3]
}

func itemToData(ctx context.Context, item *op.Item, data *OnePasswordItemResourceModel) diag.Diagnostics {
	data.ID = types.StringValue(terraformItemID(item))
	data.UUID = types.StringValue(item.ID)
	data.Vault = types.StringValue(item.Vault.ID)
	data.Title = types.StringValue(item.Title)

	for _, u := range item.URLs {
		if u.Primary {
			data.URL = types.StringValue(u.URL)
		}
	}

	tags, diagnostics := types.ListValueFrom(ctx, types.StringType, item.Tags)
	if diagnostics.HasError() {
		return diagnostics
	}
	data.Tags = tags

	data.Category = types.StringValue(string(item.Category))

	dataSections := data.Section
	for _, s := range item.Sections {
		section := OnePasswordItemResourceSectionModel{}
		newSection := true

		for i := range dataSections {
			existingID := dataSections[i].ID.ValueString()
			existingLabel := dataSections[i].Label.ValueString()
			if (s.ID != "" && s.ID == existingID) || s.Label == existingLabel {
				section = dataSections[i]
				newSection = false
			}
		}

		section.ID = types.StringValue(s.ID)
		section.Label = types.StringValue(s.Label)

		var existingFields []OnePasswordItemResourceFieldModel
		if section.Field != nil {
			existingFields = section.Field
		}
		for _, f := range item.Fields {
			if f.Section != nil && f.Section.ID == s.ID {
				dataField := OnePasswordItemResourceFieldModel{}
				newField := true

				for i := range existingFields {
					existingID := existingFields[i].ID.ValueString()
					existingLabel := existingFields[i].Label.ValueString()

					if (f.ID != "" && f.ID == existingID) || f.Label == existingLabel {
						dataField = existingFields[i]
						newField = false
					}
				}

				dataField = OnePasswordItemResourceFieldModel{
					OnePasswordItemFieldModel: OnePasswordItemFieldModel{
						ID:      types.StringValue(f.ID),
						Label:   types.StringValue(f.Label),
						Purpose: types.StringValue(string(f.Purpose)),
						Type:    types.StringValue(string(f.Type)),
						Value:   types.StringValue(f.Value),
					},
				}

				//dataField.ID = types.StringValue(f.ID)
				//dataField.Label = types.StringValue(f.Label)
				//dataField.Purpose = types.StringValue(string(f.Purpose))
				//dataField.Type = types.StringValue(string(f.Type))
				//dataField.Value = types.StringValue(f.Value)

				if f.Type == op.FieldTypeDate {
					date, err := util.SecondsToYYYYMMDD(f.Value)
					if err != nil {
						return diag.Diagnostics{diag.NewErrorDiagnostic(
							"Error parsing data",
							fmt.Sprintf("Failed to parse date value, got error: %s", err),
						)}
					}
					dataField.Value = types.StringValue(date)
				}

				if f.Recipe != nil {
					charSets := map[string]bool{}
					for _, s := range f.Recipe.CharacterSets {
						charSets[strings.ToLower(s)] = true
					}

					dataField.Recipe = PasswordRecipeModel{
						Length:  types.Int64Value(int64(f.Recipe.Length)),
						Letters: types.BoolValue(charSets["letters"]),
						Digits:  types.BoolValue(charSets["digits"]),
						Symbols: types.BoolValue(charSets["symbols"]),
					}
				}

				if newField {
					existingFields = append(existingFields, dataField)
				}
			}
		}
		section.Field = existingFields

		if newSection {
			dataSections = append(dataSections, section)
		}
	}

	data.Section = dataSections

	for _, f := range item.Fields {
		switch f.Purpose {
		case "USERNAME":
			data.Username = types.StringValue(f.Value)
		case "PASSWORD":
			data.Password = types.StringValue(f.Value)
		// TODO: Uncomment this if we decide to include note_value attribute in the resource schema.
		//case "NOTES":
		//	data.NoteValue = types.StringValue(f.Value)
		default:
			if f.Section == nil {
				// TODO: add rest of supported cases for fields with no sections
				//	data.f.Label), f.Value)
			}
		}
	}

	return nil
}
