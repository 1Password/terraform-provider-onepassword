package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	opssh "github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/ssh"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ ephemeral.EphemeralResourceWithConfigure = &OnePasswordItemEphemeral{}

func NewOnePasswordItemEphemeral() ephemeral.EphemeralResource {
	return &OnePasswordItemEphemeral{}
}

// OnePasswordItemEphemeral defines the ephemeral resource implementation.
type OnePasswordItemEphemeral struct {
	client onepassword.Client
}

// OnePasswordItemEphemeralModel describes the ephemeral resource data model.
// Fields mirror OnePasswordItemDataSourceModel so users can substitute an
// ephemeral block in place of a data source without losing access to any
// item field. See issue #330.
type OnePasswordItemEphemeralModel struct {
	ID                types.String                              `tfsdk:"id"`
	Vault             types.String                              `tfsdk:"vault"`
	UUID              types.String                              `tfsdk:"uuid"`
	Title             types.String                              `tfsdk:"title"`
	Category          types.String                              `tfsdk:"category"`
	URL               types.String                              `tfsdk:"url"`
	Hostname          types.String                              `tfsdk:"hostname"`
	Database          types.String                              `tfsdk:"database"`
	Port              types.String                              `tfsdk:"port"`
	Type              types.String                              `tfsdk:"type"`
	Tags              types.List                                `tfsdk:"tags"`
	Username          types.String                              `tfsdk:"username"`
	Password          types.String                              `tfsdk:"password"`
	NoteValue         types.String                              `tfsdk:"note_value"`
	Credential        types.String                              `tfsdk:"credential"`
	ValidFrom         types.String                              `tfsdk:"valid_from"`
	Filename          types.String                              `tfsdk:"filename"`
	PublicKey         types.String                              `tfsdk:"public_key"`
	PrivateKey        types.String                              `tfsdk:"private_key"`
	PrivateKeyOpenSSH types.String                              `tfsdk:"private_key_openssh"`
	SectionList       []OnePasswordItemSectionListModel         `tfsdk:"section"`
	SectionMap        map[string]OnePasswordItemSectionMapModel `tfsdk:"section_map"`
	File              []OnePasswordItemFileListModel            `tfsdk:"file"`
}

func (r *OnePasswordItemEphemeral) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *OnePasswordItemEphemeral) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	// Note: the ephemeral framework does not permit Computed-only blocks
	// (block counts in the result must match block counts in config), so
	// section/file are exposed as Computed ListNestedAttributes here even
	// though the data source uses blocks. The resulting access pattern in
	// HCL is identical (e.g. ephemeral.onepassword_item.x.section[0].field[0].value).
	fileNestedObject := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: fileIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: fileNameDescription,
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: fileContentDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"content_base64": schema.StringAttribute{
				MarkdownDescription: fileContentBase64Description,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: itemEphemeralDescription,

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
				MarkdownDescription: itemLookupUUIDDescription,
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
				MarkdownDescription: itemLookupTitleDescription,
				Optional:            true,
				Computed:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(enumDescription, categoryDescription, dataSourceCategories),
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
				MarkdownDescription: typeDescription,
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
				Sensitive:           true,
			},
			"credential": schema.StringAttribute{
				MarkdownDescription: credentialDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"valid_from": schema.StringAttribute{
				MarkdownDescription: validFromDescription,
				Computed:            true,
			},
			"filename": schema.StringAttribute{
				MarkdownDescription: filenameDescription,
				Computed:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: publicKeyDescription,
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: privateKeyDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"private_key_openssh": schema.StringAttribute{
				MarkdownDescription: privateKeyOpenSSHDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"section_map": schema.MapNestedAttribute{
				MarkdownDescription: sectionMapDescription,
				Computed:            true,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: sectionIDDescription,
							Computed:            true,
						},
						"field_map": schema.MapNestedAttribute{
							MarkdownDescription: fieldMapDescription,
							Computed:            true,
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fieldIDDescription,
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
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
						"file_map": schema.MapNestedAttribute{
							MarkdownDescription: fileMapDescription,
							Computed:            true,
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fileIDDescription,
										Computed:            true,
									},
									"content": schema.StringAttribute{
										MarkdownDescription: fileContentDescription,
										Computed:            true,
										Sensitive:           true,
									},
									"content_base64": schema.StringAttribute{
										MarkdownDescription: fileContentBase64Description,
										Computed:            true,
										Sensitive:           true,
									},
								},
							},
						},
					},
				},
			},
			"section": schema.ListNestedAttribute{
				MarkdownDescription: sectionListDescription,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: sectionIDDescription,
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: sectionLabelDescription,
							Computed:            true,
						},
						"field": schema.ListNestedAttribute{
							MarkdownDescription: fieldListDescription,
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fieldIDDescription,
										Computed:            true,
									},
									"label": schema.StringAttribute{
										MarkdownDescription: fieldLabelDescription,
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
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
						"file": schema.ListNestedAttribute{
							MarkdownDescription: fileListDescription,
							Computed:            true,
							NestedObject:        fileNestedObject,
						},
					},
				},
			},
			"file": schema.ListNestedAttribute{
				MarkdownDescription: documentFileListDescription,
				Computed:            true,
				NestedObject:        fileNestedObject,
			},
		},
	}
}

func (r *OnePasswordItemEphemeral) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(onepassword.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected onepassword.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OnePasswordItemEphemeral) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data OnePasswordItemEphemeralModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := getItem(ctx, r.client, data.Vault.ValueString(), data.Title.ValueString(), data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read item, got error: %s", err))
		return
	}

	data.ID = types.StringValue(itemTerraformID(item))
	data.UUID = types.StringValue(item.ID)
	data.Vault = types.StringValue(item.VaultID)
	data.Title = types.StringValue(item.Title)
	data.Category = types.StringValue(strings.ToLower(string(item.Category)))

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

	for _, s := range item.Sections {
		section := OnePasswordItemSectionListModel{
			ID:    types.StringValue(s.ID),
			Label: types.StringValue(s.Label),
		}

		for _, f := range item.Fields {
			if f.SectionID != "" && f.SectionID == s.ID {
				section.Field = append(section.Field, OnePasswordItemFieldListModel{
					ID:    types.StringValue(f.ID),
					Label: types.StringValue(f.Label),
					Type:  types.StringValue(string(f.Type)),
					Value: types.StringValue(f.Value),
				})
			}
		}

		for _, f := range item.Files {
			if f.SectionID != "" && f.SectionID == s.ID {
				content, err := f.Content()
				if err != nil {
					// content has not yet been loaded, fetch it
					content, err = r.client.GetFileContent(ctx, &f, item.ID, item.VaultID)
				}
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
				}
				file := OnePasswordItemFileListModel{
					ID:            types.StringValue(f.ID),
					Name:          types.StringValue(f.Name),
					Content:       types.StringValue(string(content)),
					ContentBase64: types.StringValue(base64.StdEncoding.EncodeToString(content)),
				}
				section.File = append(section.File, file)
			}
		}

		data.SectionList = append(data.SectionList, section)
	}

	sectionMap, diag := buildSectionMap(ctx, item, r.client)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.SectionMap = sectionMap

	for _, f := range item.Fields {
		switch f.Purpose {
		case model.FieldPurposeUsername:
			data.Username = types.StringValue(f.Value)
		case model.FieldPurposePassword:
			data.Password = types.StringValue(f.Value)
		case model.FieldPurposeNotes:
			data.NoteValue = types.StringValue(f.Value)
		default:
			if f.SectionID == "" {
				switch f.ID {
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
				case "type", "database_type":
					data.Type = types.StringValue(f.Value)
				case "public_key":
					data.PublicKey = types.StringValue(f.Value)
				case "private_key":
					data.PrivateKey = types.StringValue(f.Value)
					openSSHPrivateKey, err := opssh.PrivateKeyToOpenSSH([]byte(f.Value), item.ID)
					if err != nil {
						resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert private key to OpenSSH format, got error: %s", err))
					}
					data.PrivateKeyOpenSSH = types.StringValue(openSSHPrivateKey)
				case "credential":
					data.Credential = types.StringValue(f.Value)
				case "validFrom":
					data.ValidFrom = types.StringValue(f.Value)
				case "filename":
					data.Filename = types.StringValue(f.Value)
				}
			}
		}
	}

	for _, f := range item.Files {
		if f.SectionID == "" {
			content, err := f.Content()
			if err != nil {
				// content has not yet been loaded, fetch it
				content, err = r.client.GetFileContent(ctx, &f, item.ID, item.VaultID)
			}
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
			}
			file := OnePasswordItemFileListModel{
				ID:            types.StringValue(f.ID),
				Name:          types.StringValue(f.Name),
				Content:       types.StringValue(string(content)),
				ContentBase64: types.StringValue(base64.StdEncoding.EncodeToString(content)),
			}
			data.File = append(data.File, file)
		}
	}

	// Save data into ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
