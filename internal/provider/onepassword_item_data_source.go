// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/internal/onepassword"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OnePasswordItemDataSource{}

func NewOnePasswordItemDataSource() datasource.DataSource {
	return &OnePasswordItemDataSource{}
}

// OnePasswordItemDataSource defines the data source implementation.
type OnePasswordItemDataSource struct {
	client *http.Client
}

// ExampleDataSourceModel describes the data source data model.
type ExampleDataSourceModel struct {
	ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	Id                    types.String `tfsdk:"id"`
}

func (d *OnePasswordItemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (d *OnePasswordItemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Use this data source to get details of an item by its vault uuid and either the title or the uuid of the item.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Terraform resource identifier for this item in the format `vaults/<vault_id>/items/<item_id>`",
				Computed:            true,
			},
			"vault": schema.StringAttribute{
				MarkdownDescription: vaultUUIDDescription,
				Required:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the item to retrieve. This field will be populated with the UUID of the item if the item it looked up by its title.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The title of the item to retrieve. This field will be populated with the title of the item if the item it looked up by its UUID.",
				Optional:            true,
				Computed:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(enumDescription, categoryDescription, categories),
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: urlDescription,
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: dbHostnameDescription,
				Computed:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: dbDatabaseDescription,
				Computed:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: dbPortDescription,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(enumDescription, dbTypeDescription, dbTypes),
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: tagsDescription,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: usernameDescription,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: passwordDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"note_value": schema.StringAttribute{
				MarkdownDescription: noteValueDescription,
				Computed:            true,
				Optional:            true,
				Sensitive:           true,
			},
			//"section": schema.ListNestedAttribute{
			//	MarkdownDescription: sectionDescription,
			//	Computed:            true,
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: map[string]schema.Attribute{
			//			"id": schema.StringAttribute{
			//				MarkdownDescription: sectionIDDescription,
			//				Computed:            true,
			//			},
			//			"label": schema.StringAttribute{
			//				MarkdownDescription: sectionLabelDescription,
			//				Computed:            true,
			//			},
			//			"field": schema.ListNestedAttribute{
			//				MarkdownDescription: sectionFieldsDescription,
			//				Computed:            true,
			//				NestedObject: schema.NestedAttributeObject{
			//					Attributes: map[string]schema.Attribute{
			//						"id": schema.StringAttribute{
			//							MarkdownDescription: fieldIDDescription,
			//							Computed:            true,
			//						},
			//						"label": schema.StringAttribute{
			//							MarkdownDescription: fieldLabelDescription,
			//							Computed:            true,
			//						},
			//						"purpose": schema.StringAttribute{
			//							MarkdownDescription: fieldPurposeDescription,
			//							Computed:            true,
			//						},
			//						"type": schema.StringAttribute{
			//							MarkdownDescription: fieldTypeDescription,
			//							Computed:            true,
			//						},
			//						"value": schema.StringAttribute{
			//							MarkdownDescription: fieldValueDescription,
			//							Computed:            true,
			//							Sensitive:           true,
			//						},
			//					},
			//				},
			//			},
			//		},
			//	},
			//},
		},
		Blocks: map[string]schema.Block{
			"section": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: sectionIDDescription,
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: sectionLabelDescription,
							Computed:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"field": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fieldIDDescription,
										Computed:            true,
									},
									"label": schema.StringAttribute{
										MarkdownDescription: fieldLabelDescription,
										Computed:            true,
									},
									"purpose": schema.StringAttribute{
										MarkdownDescription: fieldPurposeDescription,
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: fieldTypeDescription,
										Computed:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: fieldValueDescription,
										Computed:            true,
										Sensitive:           true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *OnePasswordItemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OnePasswordItemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExampleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
