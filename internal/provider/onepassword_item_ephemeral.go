package provider

import (
	"context"
	"fmt"

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

// OnePasswordItemEphemeralModel describes the data source data model.
type OnePasswordItemEphemeralModel struct {
	ID                types.String `tfsdk:"id"`
	Vault             types.String `tfsdk:"vault"`
	UUID              types.String `tfsdk:"uuid"`
	Title             types.String `tfsdk:"title"`
	URL               types.String `tfsdk:"url"`
	Hostname          types.String `tfsdk:"hostname"`
	Database          types.String `tfsdk:"database"`
	Port              types.String `tfsdk:"port"`
	Type              types.String `tfsdk:"type"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	NoteValue         types.String `tfsdk:"note_value"`
	Credential        types.String `tfsdk:"credential"`
	PublicKey         types.String `tfsdk:"public_key"`
	PrivateKey        types.String `tfsdk:"private_key"`
	PrivateKeyOpenSSH types.String `tfsdk:"private_key_openssh"`
}

func (r *OnePasswordItemEphemeral) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *OnePasswordItemEphemeral) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
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

	for _, u := range item.URLs {
		if u.Primary {
			data.URL = types.StringValue(u.URL)
		}
	}

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
				}
			}
		}
	}

	// Save data into ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
