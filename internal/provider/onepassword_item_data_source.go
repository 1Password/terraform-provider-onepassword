package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	client onepassword.Client
}

// OnePasswordItemDataSourceModel describes the data source data model.
type OnePasswordItemDataSourceModel struct {
	ID        types.String                  `tfsdk:"id"`
	Vault     types.String                  `tfsdk:"vault"`
	UUID      types.String                  `tfsdk:"uuid"`
	Title     types.String                  `tfsdk:"title"`
	Category  types.String                  `tfsdk:"category"`
	URL       types.String                  `tfsdk:"url"`
	Hostname  types.String                  `tfsdk:"hostname"`
	Database  types.String                  `tfsdk:"database"`
	Port      types.String                  `tfsdk:"port"`
	Type      types.String                  `tfsdk:"type"`
	Tags      types.List                    `tfsdk:"tags"`
	Username  types.String                  `tfsdk:"username"`
	Password  types.String                  `tfsdk:"password"`
	NoteValue types.String                  `tfsdk:"note_value"`
	Section   []OnePasswordItemSectionModel `tfsdk:"section"`
}

type OnePasswordItemSectionModel struct {
	ID    types.String                `tfsdk:"id"`
	Label types.String                `tfsdk:"label"`
	Field []OnePasswordItemFieldModel `tfsdk:"field"`
}

type OnePasswordItemFieldModel struct {
	ID      types.String `tfsdk:"id"`
	Label   types.String `tfsdk:"label"`
	Purpose types.String `tfsdk:"purpose"`
	Type    types.String `tfsdk:"type"`
	Value   types.String `tfsdk:"value"`
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
				MarkdownDescription: terraformItemIDDescription,
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
						path.MatchRoot("title"),
						path.MatchRoot("uuid"),
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
				//Sensitive:           true,
			},
			"note_value": schema.StringAttribute{
				MarkdownDescription: noteValueDescription,
				Computed:            true,
				Optional:            true,
				//Sensitive:           true,
			},
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
										//Sensitive:           true,
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

func (d *OnePasswordItemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnePasswordItemDataSourceModel

	tflog.Debug(ctx, fmt.Sprintf("something %v", req.Config))
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	item, err := getItemForDataSource(ctx, d.client, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read item, got error: %s", err))
		return
	}

	data.ID = types.StringValue(itemTerraformID(item))
	data.UUID = types.StringValue(item.ID)
	data.Vault = types.StringValue(item.Vault.ID)
	data.Title = types.StringValue(item.Title)

	for _, u := range item.URLs {
		if u.Primary {
			data.URL = types.StringValue(u.URL)
		}
	}

	tags, diag := types.ListValueFrom(ctx, types.StringType, item.Tags)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tags

	data.Category = types.StringValue(strings.ToLower(string(item.Category)))

	for _, s := range item.Sections {
		section := OnePasswordItemSectionModel{
			ID:    types.StringValue(s.ID),
			Label: types.StringValue(s.Label),
		}

		for _, f := range item.Fields {
			if f.Section != nil && f.Section.ID == s.ID {
				section.Field = append(section.Field, OnePasswordItemFieldModel{
					ID:      types.StringValue(f.ID),
					Label:   types.StringValue(f.Label),
					Purpose: types.StringValue(string(f.Purpose)),
					Type:    types.StringValue(string(f.Type)),
					Value:   types.StringValue(f.Value),
				})
			}
		}

		data.Section = append(data.Section, section)
	}

	for _, f := range item.Fields {
		switch f.Purpose {
		case op.FieldPurposeUsername:
			data.Username = types.StringValue(f.Value)
		case op.FieldPurposePassword:
			data.Password = types.StringValue(f.Value)
		case op.FieldPurposeNotes:
			data.NoteValue = types.StringValue(f.Value)
		default:
			if f.Section == nil {
				switch f.Label {
				case "username":
					data.Username = types.StringValue(f.Value)
				case "password":
					data.Password = types.StringValue(f.Value)
				case "hostname", "server":
					data.Hostname = types.StringValue(f.Value)
				case "database":
					data.Database = types.StringValue(f.Value)
				case "port":
					data.Port = types.StringValue(f.Value)
				case "type":
					data.Type = types.StringValue(f.Value)
				}
			}
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read an item data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func getItemForDataSource(ctx context.Context, client onepassword.Client, data OnePasswordItemDataSourceModel) (*op.Item, error) {
	vaultUUID := data.Vault.ValueString()
	itemTitle := data.Title.ValueString()
	itemUUID := data.UUID.ValueString()

	if itemTitle != "" {
		return client.GetItemByTitle(ctx, itemTitle, vaultUUID)
	}
	if itemUUID != "" {
		return client.GetItem(ctx, itemUUID, vaultUUID)
	}
	return nil, errors.New("uuid or title must be set")
}
