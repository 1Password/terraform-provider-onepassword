package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

var _ datasource.DataSource = &OnePasswordItemsDataSource{}

func NewOnePasswordItemsDataSource() datasource.DataSource {
	return &OnePasswordItemsDataSource{}
}

type OnePasswordItemsDataSource struct {
	client onepassword.Client
}

type OnePasswordItemsDataSourceModel struct {
	ID          types.String                                  `tfsdk:"id"`
	Vault       types.String                                  `tfsdk:"vault"`
	Titles      []types.String                                `tfsdk:"titles"`
	Items       map[string]OnePasswordItemsEntryModel         `tfsdk:"items"`
	Credentials map[string]types.String                       `tfsdk:"credentials"`
}

type OnePasswordItemsEntryModel struct {
	ID         types.String `tfsdk:"id"`
	Title      types.String `tfsdk:"title"`
	Category   types.String `tfsdk:"category"`
	Credential types.String `tfsdk:"credential"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	NoteValue  types.String `tfsdk:"note_value"`
	URL        types.String `tfsdk:"url"`
	Tags       types.List   `tfsdk:"tags"`
}

func (d *OnePasswordItemsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_items"
}

func (d *OnePasswordItemsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this to batch-read multiple items from a vault. Returns item details and a convenience credentials map.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: terraformItemIDDescription,
				Computed:            true,
			},
			"vault": schema.StringAttribute{
				MarkdownDescription: vaultUUIDDescription,
				Required:            true,
			},
			"titles": schema.ListAttribute{
				MarkdownDescription: "A list of item titles (or UUIDs) to retrieve from the vault.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"items": schema.MapNestedAttribute{
				MarkdownDescription: "A map from item title to item details.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: itemUUIDDescription,
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: itemTitleDescription,
							Computed:            true,
						},
						"category": schema.StringAttribute{
							MarkdownDescription: categoryDescription,
							Computed:            true,
						},
						"credential": schema.StringAttribute{
							MarkdownDescription: credentialDescription,
							Computed:            true,
							Sensitive:           true,
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
						"url": schema.StringAttribute{
							MarkdownDescription: urlDescription,
							Computed:            true,
						},
						"tags": schema.ListAttribute{
							MarkdownDescription: tagsDescription,
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"credentials": schema.MapAttribute{
				MarkdownDescription: "A map from item title to its primary credential value (password or API credential). This is the most commonly needed secret value for each item.",
				Computed:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *OnePasswordItemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OnePasswordItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnePasswordItemsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	titles := make([]string, len(data.Titles))
	for i, t := range data.Titles {
		titles[i] = t.ValueString()
	}

	items, err := d.client.GetItems(ctx, data.Vault.ValueString(), titles)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read items, got error: %s", err))
		return
	}

	itemsMap := make(map[string]OnePasswordItemsEntryModel, len(items))
	credentialsMap := make(map[string]types.String, len(items))

	for _, item := range items {
		entry := OnePasswordItemsEntryModel{
			ID:       types.StringValue(item.ID),
			Title:    types.StringValue(item.Title),
			Category: types.StringValue(strings.ToLower(string(item.Category))),
		}

		var primaryURL string
		for _, u := range item.URLs {
			if u.Primary {
				primaryURL = u.URL
			}
		}
		entry.URL = types.StringValue(primaryURL)

		tags, diag := types.ListValueFrom(ctx, types.StringType, item.Tags)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		entry.Tags = tags

		var credential, username, password, noteValue string
		for _, f := range item.Fields {
			switch f.Purpose {
			case model.FieldPurposeUsername:
				username = f.Value
			case model.FieldPurposePassword:
				password = f.Value
			case model.FieldPurposeNotes:
				noteValue = f.Value
			default:
				if f.SectionID == "" {
					switch f.ID {
					case "username":
						username = f.Value
					case "password":
						password = f.Value
					case "credential":
						credential = f.Value
					}
				}
			}
		}

		entry.Username = types.StringValue(username)
		entry.Password = types.StringValue(password)
		entry.Credential = types.StringValue(credential)
		entry.NoteValue = types.StringValue(noteValue)

		itemsMap[item.Title] = entry

		// credentials map: prefer credential (API cred), fall back to password
		if credential != "" {
			credentialsMap[item.Title] = types.StringValue(credential)
		} else {
			credentialsMap[item.Title] = types.StringValue(password)
		}
	}

	data.Items = itemsMap
	data.Credentials = credentialsMap
	data.ID = types.StringValue(fmt.Sprintf("vaults/%s/items", data.Vault.ValueString()))

	tflog.Trace(ctx, "read items data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
