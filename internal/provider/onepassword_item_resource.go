package provider

import (
	"context"
	"fmt"
	"regexp"
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

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OnePasswordItemResource{}
var _ resource.ResourceWithImportState = &OnePasswordItemResource{}
var _ resource.ResourceWithValidateConfig = &OnePasswordItemResource{}

func NewOnePasswordItemResource() resource.Resource {
	return &OnePasswordItemResource{}
}

// OnePasswordItemResource defines the resource implementation.
type OnePasswordItemResource struct {
	client onepassword.Client
}

// OnePasswordItemResourceModel describes the resource data model.
type OnePasswordItemResourceModel struct {
	ID                 types.String                                      `tfsdk:"id"`
	UUID               types.String                                      `tfsdk:"uuid"`
	Vault              types.String                                      `tfsdk:"vault"`
	Category           types.String                                      `tfsdk:"category"`
	Title              types.String                                      `tfsdk:"title"`
	URL                types.String                                      `tfsdk:"url"`
	Hostname           types.String                                      `tfsdk:"hostname"`
	Database           types.String                                      `tfsdk:"database"`
	Port               types.String                                      `tfsdk:"port"`
	Type               types.String                                      `tfsdk:"type"`
	Tags               types.List                                        `tfsdk:"tags"`
	Username           types.String                                      `tfsdk:"username"`
	Password           types.String                                      `tfsdk:"password"`
	PasswordWO         types.String                                      `tfsdk:"password_wo"`
	PasswordWOVersion  types.Int64                                       `tfsdk:"password_wo_version"`
	NoteValue          types.String                                      `tfsdk:"note_value"`
	NoteValueWO        types.String                                      `tfsdk:"note_value_wo"`
	NoteValueWOVersion types.Int64                                       `tfsdk:"note_value_wo_version"`
	SectionList        []OnePasswordItemResourceSectionListModel         `tfsdk:"section"`
	SectionMap         map[string]OnePasswordItemResourceSectionMapModel `tfsdk:"section_map"`
	Recipe             []PasswordRecipeModel                             `tfsdk:"password_recipe"`
}

type PasswordRecipeModel struct {
	Length  types.Int64 `tfsdk:"length"`
	Digits  types.Bool  `tfsdk:"digits"`
	Symbols types.Bool  `tfsdk:"symbols"`
}

// OnePasswordItemResourceSectionListModel is used for list-based sections
type OnePasswordItemResourceSectionListModel struct {
	ID        types.String                        `tfsdk:"id"`
	Label     types.String                        `tfsdk:"label"`
	FieldList []OnePasswordItemResourceFieldModel `tfsdk:"field"`
}

// OnePasswordItemResourceSectionMapModel is used for map-based sections
// The map key serves as the section label
type OnePasswordItemResourceSectionMapModel struct {
	ID       types.String                                    `tfsdk:"id"`
	FieldMap map[string]OnePasswordItemResourceFieldMapModel `tfsdk:"field_map"`
}

// OnePasswordItemResourceFieldMapModel is used for map-based fields (field_map attribute)
// The map key serves as the field label
type OnePasswordItemResourceFieldMapModel struct {
	ID     types.String         `tfsdk:"id"`
	Type   types.String         `tfsdk:"type"`
	Value  types.String         `tfsdk:"value"`
	Recipe *PasswordRecipeModel `tfsdk:"password_recipe"`
}

type OnePasswordItemResourceFieldModel struct {
	ID     types.String          `tfsdk:"id"`
	Label  types.String          `tfsdk:"label"`
	Type   types.String          `tfsdk:"type"`
	Value  types.String          `tfsdk:"value"`
	Recipe []PasswordRecipeModel `tfsdk:"password_recipe"`
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

	sectionNestedObjectSchemaForMap := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: sectionIDDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseNonNullStateForUnknown(),
				},
			},
			"field_map": schema.MapNestedAttribute{
				MarkdownDescription: fieldMapDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: fieldIDDescription,
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("STRING"),
							Validators: []validator.String{
								stringvalidator.OneOf(fieldTypes...),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: fieldValueDescription,
							Optional:            true,
							Computed:            true,
							Sensitive:           true,
							PlanModifiers: []planmodifier.String{
								PasswordValueModifierForMapField(),
							},
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("password_recipe"),
								),
								validateMonthYear(),
							},
						},
						"password_recipe": schema.SingleNestedAttribute{
							MarkdownDescription: passwordRecipeDescription,
							Optional:            true,
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
					},
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
			"note_value_wo": schema.StringAttribute{
				MarkdownDescription: noteValueWriteOnceDescription,
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.Expressions{path.MatchRoot("note_value")}...,
					),
					stringvalidator.AlsoRequires(
						path.Expressions{path.MatchRoot("note_value_wo_version")}...,
					),
				},
			},
			"note_value_wo_version": schema.Int64Attribute{
				MarkdownDescription: noteValueWriteOnceVersionDescription,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(
						path.Expressions{path.MatchRoot("note_value")}...,
					),
					int64validator.AlsoRequires(
						path.Expressions{path.MatchRoot("note_value_wo")}...,
					),
				},
			},
			"section_map": schema.MapNestedAttribute{
				MarkdownDescription: sectionMapDescription,
				Optional:            true,
				NestedObject:        sectionNestedObjectSchemaForMap,
			},
		},
		Blocks: map[string]schema.Block{
			"section": schema.ListNestedBlock{
				MarkdownDescription: sectionListDescription,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: sectionIDDescription,
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
						},
						"label": schema.StringAttribute{
							MarkdownDescription: sectionLabelDescription,
							Required:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"field": schema.ListNestedBlock{
							MarkdownDescription: fieldListDescription,
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: fieldIDDescription,
										Optional:            true,
										Computed:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseNonNullStateForUnknown(),
										},
									},
									"label": schema.StringAttribute{
										MarkdownDescription: fieldLabelDescription,
										Required:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("STRING"),
										Validators: []validator.String{
											stringvalidator.OneOf(fieldTypes...),
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
										Validators: []validator.String{
											stringvalidator.ConflictsWith(
												path.MatchRelative().AtParent().AtName("password_recipe"),
											),
											validateMonthYear(),
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

func (r *OnePasswordItemResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var sectionList types.List
	var sectionMap types.Map

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("section"), &sectionList)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("section_map"), &sectionMap)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If either is unknown (e.g., dynamic block with unknown for_each), skip validation
	// Validation will happen later when values are known
	if sectionList.IsUnknown() || sectionMap.IsUnknown() {
		return
	}

	// Check if both are set (non-null and have elements)
	hasSectionList := !sectionList.IsNull() && len(sectionList.Elements()) > 0
	hasSectionMap := !sectionMap.IsNull() && len(sectionMap.Elements()) > 0

	if hasSectionMap && hasSectionList {
		resp.Diagnostics.AddError(
			"Conflicting Section Definitions",
			"Cannot use both 'section' (list) and 'section_map' (map) at the same time. Please use only one of them.",
		)
	}
}

func (r *OnePasswordItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnePasswordItemResourceModel
	var config OnePasswordItemResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the config to get the original password_wo value as it's not stored nor inside plan neither the state.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle write-only fields
	handleWriteOnlyField(config.PasswordWOVersion, config.PasswordWO, &plan.Password)
	handleWriteOnlyField(config.NoteValueWOVersion, config.NoteValueWO, &plan.NoteValue)

	item, diagnostics := stateToModel(ctx, plan)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdItem, err := r.client.CreateItem(ctx, item, item.VaultID)
	if err != nil {
		resp.Diagnostics.AddError("1Password Item create error", fmt.Sprintf("Error creating 1Password item, got error %s", err))
		return
	}

	resp.Diagnostics.Append(modelToState(ctx, createdItem, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once created, clear write-only fields from state
	clearWriteOnlyFieldFromState(config.PasswordWOVersion, &plan.Password)
	clearWriteOnlyFieldFromState(config.NoteValueWOVersion, &plan.NoteValue)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnePasswordItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnePasswordItemResourceModel

	// Read Terraform prior state state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vaultUUID, itemUUID := vaultAndItemUUID(state.ID.ValueString())
	item, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
	if err != nil {
		// If the resource no longer exists, remove it from state
		// The next Terraform plan will recreate the resource
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("1Password Item read error", fmt.Sprintf("Could not get item '%s' from vault '%s', got error: %s", itemUUID, vaultUUID, err))
		return
	}

	resp.Diagnostics.Append(modelToState(ctx, item, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once read, clear write-only fields from state
	clearWriteOnlyFieldFromState(state.PasswordWOVersion, &state.Password)
	clearWriteOnlyFieldFromState(state.NoteValueWOVersion, &state.NoteValue)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnePasswordItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnePasswordItemResourceModel
	var config OnePasswordItemResourceModel
	var state OnePasswordItemResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	// Handle all write-only fields
	vaultUUID, itemUUID := vaultAndItemUUID(plan.ID.ValueString())
	err := handleWriteOnlyFieldUpdates(&config, &state, &plan, func() (*model.Item, error) {
		item, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
		if err != nil {
			return nil, fmt.Errorf("could not read item '%s' from vault '%s' to preserve write-only fields: %s", itemUUID, vaultUUID, err)
		}
		return item, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("1Password Item read error", err.Error())
		return
	}

	item, diagnostics := stateToModel(ctx, plan)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedItem, err := r.client.UpdateItem(ctx, item, plan.Vault.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item update error", fmt.Sprintf("Could not update item '%s' from vault '%s', got error: %s", plan.UUID.ValueString(), plan.Vault.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(modelToState(ctx, updatedItem, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once updated, always clear write-only fields from state
	clearWriteOnlyFieldFromState(config.PasswordWOVersion, &plan.Password)
	clearWriteOnlyFieldFromState(config.NoteValueWOVersion, &plan.NoteValue)

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnePasswordItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnePasswordItemResourceModel

	// Read Terraform prior state state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	item, diagnostics := stateToModel(ctx, state)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteItem(ctx, item, state.Vault.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item delete error", fmt.Sprintf("Could not delete item '%s' from vault '%s', got error: %s", state.UUID.ValueString(), state.Vault.ValueString(), err))
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

// isNotFoundError checks if an error indicates that a resource was not found.
// Different client implementations return different error when item is not found:
//   - Connect: "status 404: item ... not found"
//   - SDK: "item couldn't be found" (when item doesn't exist)
//   - SDK: "item is not in an active state" (when item was removed)
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "404") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "item couldn't be found") ||
		strings.Contains(errMsg, "item is not in an active state")
}

func modelToState(ctx context.Context, modelItem *model.Item, state *OnePasswordItemResourceModel) diag.Diagnostics {
	state.ID = setStringValue(itemTerraformID(modelItem))
	state.UUID = setStringValue(modelItem.ID)
	state.Vault = setStringValue(modelItem.VaultID)
	state.Title = setStringValuePreservingEmpty(modelItem.Title, state.Title)
	state.Category = setStringValue(strings.ToLower(string(modelItem.Category)))

	if len(state.SectionMap) > 0 {
		state.SectionMap = toStateSectionsAndFieldsMap(modelItem, state.SectionMap)
	} else {
		state.SectionList = toStateSectionsAndFieldsList(modelItem.Sections, modelItem.Fields, state.SectionList)
	}

	toStateTopLevelFields(modelItem.Fields, state)

	for _, u := range modelItem.URLs {
		if u.Primary {
			state.URL = setStringValuePreservingEmpty(u.URL, state.URL)
		}
	}

	tags, diagnostics := toStateTags(ctx, modelItem.Tags, state.Tags)
	if diagnostics.HasError() {
		return diagnostics
	}
	state.Tags = tags

	// Password is not set for secure notes
	if modelItem.Category == model.SecureNote && state.Password.IsUnknown() {
		state.Password = types.StringNull()
	}

	return nil
}

func stateToModel(ctx context.Context, state OnePasswordItemResourceModel) (*model.Item, diag.Diagnostics) {
	modelItem := &model.Item{
		ID:      state.UUID.ValueString(),
		VaultID: state.Vault.ValueString(),
		Title:   state.Title.ValueString(),
		URLs: []model.ItemURL{
			{
				Primary: true,
				URL:     state.URL.ValueString(),
			},
		},
	}

	password := state.Password.ValueString()
	recipe, err := parseGeneratorRecipeList(state.Recipe)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Error parsing generator recipe",
			fmt.Sprintf("Failed to parse generator recipe, got error: %s", err),
		)}
	}

	switch state.Category.ValueString() {
	case "login":
		modelItem.Category = model.Login
		modelItem.Fields = toModelLoginFields(state, password, recipe)
	case "password":
		modelItem.Category = model.Password
		modelItem.Fields = toModelPasswordFields(state, password, recipe)
	case "database":
		modelItem.Category = model.Database
		modelItem.Fields = toModelDatabaseFields(state, password, recipe)
	case "secure_note":
		modelItem.Category = model.SecureNote
		modelItem.Fields = toModelSecureNoteFields(state)
	}

	tags, diagnostics := toModelTags(ctx, state)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	modelItem.Tags = tags

	if len(state.SectionMap) > 0 {
		diagnostics = toModelSectionsFromMap(state, modelItem)
	} else {
		diagnostics = toModelSections(state, modelItem)
	}

	if diagnostics.HasError() {
		return nil, diagnostics
	}

	return modelItem, nil
}

func parseGeneratorRecipeList(recipeObject []PasswordRecipeModel) (*model.GeneratorRecipe, error) {
	if len(recipeObject) == 0 {
		return nil, nil
	}

	return parseGeneratorRecipeFromModel(&recipeObject[0])
}

func addRecipe(f *model.ItemField, r *model.GeneratorRecipe) {
	f.Recipe = r

	// Check to see if the current value adheres to the recipe

	var recipeDigits, recipeSymbols bool
	hasDigits, _ := regexp.MatchString("[0-9]", f.Value)
	hasSymbols, _ := regexp.MatchString("[^a-zA-Z0-9]", f.Value)

	for _, s := range r.CharacterSets {
		switch s {
		case model.CharacterSetDigits:
			recipeDigits = true
		case model.CharacterSetSymbols:
			recipeSymbols = true
		}
	}

	if hasDigits != recipeDigits ||
		hasSymbols != recipeSymbols ||
		len(f.Value) != r.Length {
		f.Generate = true
	}
}
