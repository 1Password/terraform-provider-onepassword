package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

	"github.com/hashicorp/go-uuid"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/util"
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
	ID                types.String                          `tfsdk:"id"`
	UUID              types.String                          `tfsdk:"uuid"`
	Vault             types.String                          `tfsdk:"vault"`
	Category          types.String                          `tfsdk:"category"`
	Title             types.String                          `tfsdk:"title"`
	URL               types.String                          `tfsdk:"url"`
	Hostname          types.String                          `tfsdk:"hostname"`
	Database          types.String                          `tfsdk:"database"`
	Port              types.String                          `tfsdk:"port"`
	Type              types.String                          `tfsdk:"type"`
	Tags              types.List                            `tfsdk:"tags"`
	Username          types.String                          `tfsdk:"username"`
	Password          types.String                          `tfsdk:"password"`
	PasswordWO        types.String                          `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64                           `tfsdk:"password_wo_version"`
	NoteValue         types.String                          `tfsdk:"note_value"`
	Section           []OnePasswordItemResourceSectionModel `tfsdk:"section"`
	Recipe            []PasswordRecipeModel                 `tfsdk:"password_recipe"`
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
	ID      types.String          `tfsdk:"id"`
	Label   types.String          `tfsdk:"label"`
	Purpose types.String          `tfsdk:"purpose"`
	Type    types.String          `tfsdk:"type"`
	Value   types.String          `tfsdk:"value"`
	Recipe  []PasswordRecipeModel `tfsdk:"password_recipe"`
}

func (r *OnePasswordItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

func (r *OnePasswordItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO: Consider using SingleNested
	passwordRecipeBlockSchema := schema.ListNestedBlock{
		MarkdownDescription: passwordRecipeDescription,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"length": schema.Int64Attribute{
					MarkdownDescription: passwordLengthDescription,
					Optional:            true,
					Computed:            true,
					Default:             int64default.StaticInt64(32),
					Validators: []validator.Int64{
						int64validator.Between(1, 64),
					},
				},
				"letters": schema.BoolAttribute{
					MarkdownDescription: passwordLettersDescription,
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"digits": schema.BoolAttribute{
					MarkdownDescription: passwordDigitsDescription,
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"symbols": schema.BoolAttribute{
					MarkdownDescription: passwordSymbolsDescription,
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A 1Password Item.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: terraformItemIDDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					validateOTP(),
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: itemUUIDDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Computed:            true,
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
				//Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					ValueModifier(),
				},
			},
			"password_wo": schema.StringAttribute{
				MarkdownDescription: passwordWriteOnceDescription,
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.Expressions{path.MatchRoot("password")}...,
					),
					stringvalidator.AlsoRequires(
						path.Expressions{path.MatchRoot("password_wo_version")}...,
					),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				MarkdownDescription: passwordWriteOnceVersionDescription,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(
						path.Expressions{path.MatchRoot("password")}...,
					),
					int64validator.AlsoRequires(
						path.Expressions{path.MatchRoot("password_wo")}...,
					),
				},
			},
			"note_value": schema.StringAttribute{
				MarkdownDescription: noteValueDescription,
				Optional:            true,
				Sensitive:           true,
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
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
										Computed:            true,
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
										PlanModifiers: []planmodifier.String{
											ValueModifier(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"password_recipe": passwordRecipeBlockSchema,
								},
							},
						},
					},
				},
			},
			"password_recipe": passwordRecipeBlockSchema,
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
	var config OnePasswordItemResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the config to get the original password_wo value as it's not stored nor inside plan neither the state.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the password_wo as password for creation when wo variant is used.
	writeOnly := false
	if !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown() {
		data.Password = config.PasswordWO
		writeOnly = true
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	item, diagnostics := dataToItem(ctx, data)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdItem, err := r.client.CreateItem(ctx, item, item.Vault.ID)
	if err != nil {
		resp.Diagnostics.AddError("1Password Item create error", fmt.Sprintf("Error creating 1Password item, got error %s", err))
		return
	}

	resp.Diagnostics.Append(itemToData(ctx, createdItem, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once created, clear password from state if wo variant is used as password should never be stored
	if writeOnly {
		data.Password = types.StringNull()
	}

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

	// Check if wo variant is used based on the wo_version stored in the prior state
	writeOnly := false
	if !data.PasswordWOVersion.IsNull() {
		writeOnly = true
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	vaultUUID, itemUUID := vaultAndItemUUID(data.ID.ValueString())
	item, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
	if err != nil {
		resp.Diagnostics.AddError("1Password Item read error", fmt.Sprintf("Could not get item '%s' from vault '%s', got error: %s", itemUUID, vaultUUID, err))
		return
	}

	resp.Diagnostics.Append(itemToData(ctx, item, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once read, clear password from state if wo variant is used as password should never be stored
	if writeOnly {
		data.Password = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OnePasswordItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OnePasswordItemResourceModel
	var config OnePasswordItemResourceModel
	var state OnePasswordItemResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the config to get the current password_wo value as it's not stored nor inside plan neither the state.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the previous state to detect if the password_wo_version should trigger a password update.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure password is field with new wo value if the current config version is != from the previous state one.
	writeOnce := false
	if !config.PasswordWOVersion.IsNull() && config.PasswordWOVersion != state.PasswordWOVersion {
		data.Password = config.PasswordWO
		writeOnce = true
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	item, diagnostics := dataToItem(ctx, data)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, _ := json.Marshal(item)
	tflog.Info(ctx, "update op payload: "+string(payload))

	updatedItem, err := r.client.UpdateItem(ctx, item, data.Vault.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item update error", fmt.Sprintf("Could not update item '%s' from vault '%s', got error: %s", data.UUID.ValueString(), data.Vault.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(itemToData(ctx, updatedItem, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once updated, always clear password from state - as it should never be stored when wo variant is used.
	if writeOnce {
		data.Password = types.StringNull()
	}

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
	item, diagnostics := dataToItem(ctx, data)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteItem(ctx, item, data.Vault.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item delete error", fmt.Sprintf("Could not delete item '%s' from vault '%s', got error: %s", data.UUID.ValueString(), data.Vault.ValueString(), err))
		return
	}
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
	data.ID = setStringValue(itemTerraformID(item))
	data.UUID = setStringValue(item.ID)
	data.Vault = setStringValue(item.Vault.ID)
	data.Title = setStringValue(item.Title)

	for _, u := range item.URLs {
		if u.Primary {
			data.URL = setStringValue(u.URL)
		}
	}

	var dataTags []string
	diagnostics := data.Tags.ElementsAs(ctx, &dataTags, false)
	if diagnostics.HasError() {
		return diagnostics
	}

	sort.Strings(dataTags)
	if !reflect.DeepEqual(dataTags, item.Tags) {
		tags, diagnostics := types.ListValueFrom(ctx, types.StringType, item.Tags)
		if diagnostics.HasError() {
			return diagnostics
		}

		if item.Tags != nil || dataTags == nil {
			data.Tags = tags
		}
	}

	data.Category = setStringValue(strings.ToLower(string(item.Category)))

	dataSections := data.Section
	for _, s := range item.Sections {
		section := OnePasswordItemResourceSectionModel{}
		posSection := -1
		newSection := true

		for i := range dataSections {
			existingID := dataSections[i].ID.ValueString()
			existingLabel := dataSections[i].Label.ValueString()
			if (s.ID != "" && s.ID == existingID) || s.Label == existingLabel {
				section = dataSections[i]
				posSection = i
				newSection = false
			}
		}

		section.ID = setStringValue(s.ID)
		section.Label = setStringValue(s.Label)

		var existingFields []OnePasswordItemResourceFieldModel
		if section.Field != nil {
			existingFields = section.Field
		}
		for _, f := range item.Fields {
			if f.Section != nil && f.Section.ID == s.ID {
				dataField := OnePasswordItemResourceFieldModel{}
				posField := -1
				newField := true

				for i := range existingFields {
					existingID := existingFields[i].ID.ValueString()
					existingLabel := existingFields[i].Label.ValueString()

					if (f.ID != "" && f.ID == existingID) || f.Label == existingLabel {
						dataField = existingFields[i]
						posField = i
						newField = false
					}
				}

				dataField.ID = setStringValue(f.ID)
				dataField.Label = setStringValue(f.Label)
				dataField.Purpose = setStringValue(string(f.Purpose))
				dataField.Type = setStringValue(string(f.Type))
				dataField.Value = setStringValue(f.Value)

				if f.Type == op.FieldTypeDate {
					date, err := util.SecondsToYYYYMMDD(f.Value)
					if err != nil {
						return diag.Diagnostics{diag.NewErrorDiagnostic(
							"Error parsing data",
							fmt.Sprintf("Failed to parse date value, got error: %s", err),
						)}
					}
					dataField.Value = setStringValue(date)
				}

				if f.Recipe != nil {
					charSets := map[string]bool{}
					for _, s := range f.Recipe.CharacterSets {
						charSets[strings.ToLower(s)] = true
					}

					dataField.Recipe = []PasswordRecipeModel{{
						Length:  types.Int64Value(int64(f.Recipe.Length)),
						Letters: types.BoolValue(charSets["letters"]),
						Digits:  types.BoolValue(charSets["digits"]),
						Symbols: types.BoolValue(charSets["symbols"]),
					}}
				}

				if newField {
					existingFields = append(existingFields, dataField)
				} else {
					existingFields[posField] = dataField
				}
			}
		}
		section.Field = existingFields

		if newSection {
			dataSections = append(dataSections, section)
		} else {
			dataSections[posSection] = section
		}
	}

	data.Section = dataSections

	for _, f := range item.Fields {
		switch f.Purpose {
		case op.FieldPurposeUsername:
			data.Username = setStringValue(f.Value)
		case op.FieldPurposePassword:
			data.Password = setStringValue(f.Value)
		case op.FieldPurposeNotes:
			data.NoteValue = setStringValue(f.Value)
		default:
			if f.Section == nil {
				switch f.Label {
				case "username":
					data.Username = setStringValue(f.Value)
				case "password":
					data.Password = setStringValue(f.Value)
				case "hostname", "server":
					data.Hostname = setStringValue(f.Value)
				case "database":
					data.Database = setStringValue(f.Value)
				case "port":
					data.Port = setStringValue(f.Value)
				case "type":
					data.Type = setStringValue(f.Value)
				}
			}
		}
	}

	if item.Category == op.SecureNote && data.Password.IsUnknown() {
		data.Password = types.StringNull()
	}

	return nil
}

func dataToItem(ctx context.Context, data OnePasswordItemResourceModel) (*op.Item, diag.Diagnostics) {
	item := &op.Item{
		ID: data.UUID.ValueString(),
		Vault: op.ItemVault{
			ID: data.Vault.ValueString(),
		},
		Title: data.Title.ValueString(),
		URLs: []op.ItemURL{
			{
				Primary: true,
				URL:     data.URL.ValueString(),
			},
		},
	}

	var tags []string
	diagnostics := data.Tags.ElementsAs(ctx, &tags, false)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	item.Tags = tags

	password := data.Password.ValueString()
	recipe, err := parseGeneratorRecipe(data.Recipe)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Error parsing generator recipe",
			fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
		)}
	}

	switch data.Category.ValueString() {
	case "login":
		item.Category = op.Login
		item.Fields = []*op.ItemField{
			{
				ID:      "username",
				Label:   "username",
				Purpose: op.FieldPurposeUsername,
				Type:    op.FieldTypeString,
				Value:   data.Username.ValueString(),
			},
			{
				ID:       "password",
				Label:    "password",
				Purpose:  op.FieldPurposePassword,
				Type:     op.FieldTypeConcealed,
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
			{
				ID:      "notesPlain",
				Label:   "notesPlain",
				Type:    op.FieldTypeString,
				Purpose: op.FieldPurposeNotes,
				Value:   data.NoteValue.ValueString(),
			},
		}
	case "password":
		item.Category = op.Password
		item.Fields = []*op.ItemField{
			{
				ID:       "password",
				Label:    "password",
				Purpose:  op.FieldPurposePassword,
				Type:     op.FieldTypeConcealed,
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
			{
				ID:      "notesPlain",
				Label:   "notesPlain",
				Type:    op.FieldTypeString,
				Purpose: op.FieldPurposeNotes,
				Value:   data.NoteValue.ValueString(),
			},
		}
	case "database":
		item.Category = op.Database
		item.Fields = []*op.ItemField{
			{
				ID:    "username",
				Label: "username",
				Type:  op.FieldTypeString,
				Value: data.Username.ValueString(),
			},
			{
				ID:       "password",
				Label:    "password",
				Type:     op.FieldTypeConcealed,
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
			{
				ID:    "hostname",
				Label: "hostname",
				Type:  op.FieldTypeString,
				Value: data.Hostname.ValueString(),
			},
			{
				ID:    "database",
				Label: "database",
				Type:  op.FieldTypeString,
				Value: data.Database.ValueString(),
			},
			{
				ID:    "port",
				Label: "port",
				Type:  op.FieldTypeString,
				Value: data.Port.ValueString(),
			},
			{
				ID:    "database_type",
				Label: "type",
				Type:  op.FieldTypeMenu,
				Value: data.Type.ValueString(),
			},
			{
				ID:      "notesPlain",
				Label:   "notesPlain",
				Type:    op.FieldTypeString,
				Purpose: op.FieldPurposeNotes,
				Value:   data.NoteValue.ValueString(),
			},
		}
	case "secure_note":
		item.Category = op.SecureNote
		item.Fields = []*op.ItemField{
			{
				ID:      "notesPlain",
				Label:   "notesPlain",
				Type:    op.FieldTypeString,
				Purpose: op.FieldPurposeNotes,
				Value:   data.NoteValue.ValueString(),
			},
		}
	}

	sections := data.Section

	for _, section := range sections {
		sectionID := section.ID.ValueString()
		if sectionID == "" {
			sid, err := uuid.GenerateUUID()
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
					"Item conversion error",
					fmt.Sprintf("Unable to generate a section ID, has error: %v", err),
				)}
			}
			sectionID = sid
		}

		s := &op.ItemSection{
			ID:    sectionID,
			Label: section.Label.ValueString(),
		}
		item.Sections = append(item.Sections, s)

		sectionFields := section.Field
		for _, field := range sectionFields {
			fieldID := field.ID.ValueString()
			fieldType := op.ItemFieldType(field.Type.ValueString())
			fieldValue := field.Value.ValueString()
			if fieldType == op.FieldTypeDate {
				if !util.IsValidDateFormat(fieldValue) {
					return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
						"Item conversion error",
						fmt.Sprintf("Invalid date value provided '%s'. Should be in YYYY-MM-DD format", fieldValue),
					)}
				}
			}

			f := &op.ItemField{
				Section: s,
				ID:      fieldID,
				Type:    fieldType,
				Purpose: op.ItemFieldPurpose(field.Purpose.ValueString()),
				Label:   field.Label.ValueString(),
				Value:   fieldValue,
			}

			recipe, err := parseGeneratorRecipe(field.Recipe)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
					"Item conversion error",
					fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
				)}
			}

			if recipe != nil {
				addRecipe(f, recipe)
			}

			item.Fields = append(item.Fields, f)
		}
	}

	return item, nil
}

func parseGeneratorRecipe(recipeObject []PasswordRecipeModel) (*op.GeneratorRecipe, error) {
	if recipeObject == nil || len(recipeObject) == 0 {
		return nil, nil
	}

	recipe := recipeObject[0]

	parsed := &op.GeneratorRecipe{
		Length:        32,
		CharacterSets: []string{},
	}

	length := recipe.Length.ValueInt64()
	if length > 64 {
		return nil, fmt.Errorf("password_recipe.length must be an integer between 1 and 64")
	}

	if length > 0 {
		parsed.Length = int(length)
	}

	if recipe.Letters.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, "LETTERS")
	}
	if recipe.Digits.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, "DIGITS")
	}
	if recipe.Symbols.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, "SYMBOLS")
	}

	return parsed, nil
}

func addRecipe(f *op.ItemField, r *op.GeneratorRecipe) {
	f.Recipe = r

	// Check to see if the current value adheres to the recipe

	var recipeLetters, recipeDigits, recipeSymbols bool
	hasLetters, _ := regexp.MatchString("[a-zA-Z]", f.Value)
	hasDigits, _ := regexp.MatchString("[0-9]", f.Value)
	hasSymbols, _ := regexp.MatchString("[^a-zA-Z0-9]", f.Value)

	for _, s := range r.CharacterSets {
		switch s {
		case "LETTERS":
			recipeLetters = true
		case "DIGITS":
			recipeDigits = true
		case "SYMBOLS":
			recipeSymbols = true
		}
	}

	if hasLetters != recipeLetters ||
		hasDigits != recipeDigits ||
		hasSymbols != recipeSymbols ||
		len(f.Value) != r.Length {
		f.Generate = true
	}
}
