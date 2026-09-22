package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

var _ datasource.DataSource = &OnePasswordItemsDataSource{}

func NewOnePasswordItemsDataSource() datasource.DataSource {
	return &OnePasswordItemsDataSource{}
}

type OnePasswordItemsDataSource struct {
	client onepassword.Client
}

type OnePasswordItemsDataSourceModel struct {
	ID          types.String                          `tfsdk:"id"`
	Vault       types.String                          `tfsdk:"vault"`
	Titles      []types.String                        `tfsdk:"titles"`
	Items       map[string]OnePasswordItemsEntryModel `tfsdk:"items"`
	Credentials map[string]types.String               `tfsdk:"credentials"`
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
				MarkdownDescription: "The Terraform resource identifier for this data source in the format `vaults/<vault_id>/items`.",
				Computed:            true,
			},
			"vault": schema.StringAttribute{
				MarkdownDescription: "The UUID of the vault the items are in.",
				Required:            true,
			},
			"titles": schema.ListAttribute{
				MarkdownDescription: "The items to retrieve from the vault. Each element is either an item title or an item UUID. " +
					"Each element is also used verbatim as the key for that item in `items` and `credentials`, so a UUID here " +
					"produces a UUID key, not a title key. Elements must be unique.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
				},
			},
			"items": schema.MapNestedAttribute{
				MarkdownDescription: "The retrieved items, keyed by the string supplied in `titles` rather than by the item's own title.",
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
				MarkdownDescription: "Each item's primary secret, keyed by the string supplied in `titles`. The value is the item's " +
					"`credential` field if it has one, otherwise its `password` field. An item with neither has no entry here, so " +
					"looking it up fails rather than returning an empty string. This whole map is sensitive, including its keys, " +
					"so it cannot be used directly in `for_each`; iterate over `items` instead and index `credentials` inside.",
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
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
		if t.IsNull() || t.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("titles").AtListIndex(i),
				"Empty Item Identifier",
				"Every element of titles must be a non-empty item title or item UUID.",
			)
			continue
		}
		titles[i] = t.ValueString()
	}
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.GetItems(ctx, data.Vault.ValueString(), titles)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read items, got error: %s", err))
		return
	}

	// GetItems returns one item per requested identifier, in request order, which is what lets the
	// output maps be keyed by the identifier the practitioner actually wrote.
	if len(items) != len(titles) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Requested %d items but the 1Password client returned %d. Please report this issue to the provider developers.", len(titles), len(items)),
		)
		return
	}

	itemsMap := make(map[string]OnePasswordItemsEntryModel, len(items))
	credentialsMap := make(map[string]types.String, len(items))

	for i, item := range items {
		if item == nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("The 1Password client returned no item for %q. Please report this issue to the provider developers.", titles[i]),
			)
			return
		}

		entry := OnePasswordItemsEntryModel{
			ID:       types.StringValue(item.ID),
			Title:    types.StringValue(item.Title),
			Category: types.StringValue(strings.ToLower(string(item.Category))),
		}

		for _, u := range item.URLs {
			if u.Primary {
				entry.URL = types.StringValue(u.URL)
			}
		}

		tags, diag := types.ListValueFrom(ctx, types.StringType, item.Tags)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		entry.Tags = tags

		// Presence is tracked separately from value: an item may legitimately hold an empty
		// password, and that has to stay distinguishable from having no password at all.
		var credential, password string
		var hasCredential, hasPassword bool
		for _, f := range item.Fields {
			switch f.Purpose {
			case model.FieldPurposeUsername:
				entry.Username = types.StringValue(f.Value)
			case model.FieldPurposePassword:
				password, hasPassword = f.Value, true
			case model.FieldPurposeNotes:
				entry.NoteValue = types.StringValue(f.Value)
			default:
				if f.SectionID == "" {
					switch f.ID {
					case "username":
						entry.Username = types.StringValue(f.Value)
					case "password":
						password, hasPassword = f.Value, true
					case "credential":
						credential, hasCredential = f.Value, true
					}
				}
			}
		}

		if hasPassword {
			entry.Password = types.StringValue(password)
		}
		if hasCredential {
			entry.Credential = types.StringValue(credential)
		}

		itemsMap[titles[i]] = entry

		// An item with neither a credential nor a password gets no entry at all, so that looking
		// it up is a configuration error rather than a silently empty secret.
		switch {
		case hasCredential:
			credentialsMap[titles[i]] = types.StringValue(credential)
		case hasPassword:
			credentialsMap[titles[i]] = types.StringValue(password)
		}
	}

	data.Items = itemsMap
	data.Credentials = credentialsMap
	data.ID = types.StringValue(fmt.Sprintf("vaults/%s/items", data.Vault.ValueString()))

	tflog.Trace(ctx, "read items data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
