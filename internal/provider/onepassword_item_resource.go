package provider

import (
	"context"
	"encoding/json"
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
							MarkdownDescription: sectionFieldsDescription,
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

	// Use the password_wo as password for creation when wo version is used.
	writeOnly := !config.PasswordWOVersion.IsNull()
	if writeOnly {
		// Set password from password_wo if provided (validators ensure both are set together)
		if !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown() {
			plan.Password = config.PasswordWO
		}
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client plan and make a call using it.
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

	// Once created, clear password from state if wo variant is used as password should never be stored
	if writeOnly {
		plan.Password = types.StringNull()
	}

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client state and make a call using it.
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

	// Once read, clear password from state if write only password is used and as password should never be stored in state
	writeOnly := !state.PasswordWOVersion.IsNull()
	if writeOnly {
		state.Password = types.StringNull()
	}

	// Save updated state into Terraform state
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

	// Handle password_wo: update if version increased, preserve if version unchanged.
	if !config.PasswordWOVersion.IsNull() {
		configVersion := config.PasswordWOVersion.ValueInt64()
		stateVersion := int64(0)
		if !state.PasswordWOVersion.IsNull() {
			stateVersion = state.PasswordWOVersion.ValueInt64()
		}

		if configVersion > stateVersion {
			// Version increased (or first time using password_wo) - use new password_wo value
			plan.Password = config.PasswordWO
		} else {
			// Version unchanged or decreased - preserve existing password by reading current item
			vaultUUID, itemUUID := vaultAndItemUUID(plan.ID.ValueString())
			currentItem, err := r.client.GetItem(ctx, itemUUID, vaultUUID)
			if err != nil {
				resp.Diagnostics.AddError("1Password Item read error", fmt.Sprintf("Could not read item '%s' from vault '%s' to preserve password, got error: %s", itemUUID, vaultUUID, err))
				return
			}
			// Extract password from current item, or set to null if password field doesn't exist
			passwordFound := false
			for _, f := range currentItem.Fields {
				if f.Purpose == model.FieldPurposePassword {
					plan.Password = types.StringValue(f.Value)
					passwordFound = true
					break
				}
			}
			// password field not found (user removed it in 1Password), sync to that state
			if !passwordFound {
				plan.Password = types.StringNull()
			}
		}
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client plan and make a call using it.
	item, diagnostics := stateToModel(ctx, plan)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, _ := json.Marshal(item)
	tflog.Info(ctx, "update op payload: "+string(payload))

	updatedItem, err := r.client.UpdateItem(ctx, item, plan.Vault.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("1Password Item update error", fmt.Sprintf("Could not update item '%s' from vault '%s', got error: %s", plan.UUID.ValueString(), plan.Vault.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(modelToState(ctx, updatedItem, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Once updated, always clear password from state - as it should never be stored when wo variant is used.
	if !config.PasswordWOVersion.IsNull() {
		plan.Password = types.StringNull()
	}

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client state and make a call using it.
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
	state.Section = toStateSectionsAndFields(modelItem.Sections, modelItem.Fields, state.Section)
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
	recipe, err := parseGeneratorRecipe(state.Recipe)
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

	diagnostics = toModelSections(state, modelItem)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	return modelItem, nil
}

func parseGeneratorRecipe(recipeObject []PasswordRecipeModel) (*model.GeneratorRecipe, error) {
	if len(recipeObject) == 0 {
		return nil, nil
	}

	recipe := recipeObject[0]

	parsed := &model.GeneratorRecipe{
		Length:        32,
		CharacterSets: []model.CharacterSet{},
	}

	length := recipe.Length.ValueInt64()
	if length > 64 {
		return nil, fmt.Errorf("password_recipe.length must be an integer between 1 and 64")
	}

	if length > 0 {
		parsed.Length = int(length)
	}

	if recipe.Digits.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetDigits)
	}
	if recipe.Symbols.ValueBool() {
		parsed.CharacterSets = append(parsed.CharacterSets, model.CharacterSetSymbols)
	}

	return parsed, nil
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
