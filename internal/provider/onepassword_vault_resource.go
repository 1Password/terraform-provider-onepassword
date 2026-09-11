package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

var _ resource.Resource = &OnePasswordVaultResource{}
var _ resource.ResourceWithImportState = &OnePasswordVaultResource{}
var _ resource.ResourceWithValidateConfig = &OnePasswordVaultResource{}

var vaultPermissionValues = map[string]uint32{
	"no_access":            model.VaultPermissionNoAccess,
	"manage_vault":         model.VaultPermissionManageVault,
	"reveal_item_password": model.VaultPermissionRevealItemPassword,
	"read_items":           model.VaultPermissionReadItems,
	"update_items":         model.VaultPermissionUpdateItems,
	"create_items":         model.VaultPermissionCreateItems,
	"archive_items":        model.VaultPermissionArchiveItems,
	"delete_items":         model.VaultPermissionDeleteItems,
	"update_item_history":  model.VaultPermissionUpdateItemHistory,
	"send_items":           model.VaultPermissionSendItems,
	"import_items":         model.VaultPermissionImportItems,
	"export_items":         model.VaultPermissionExportItems,
	"print_items":          model.VaultPermissionPrintItems,
}

var vaultPermissionNames = func() []string {
	names := make([]string, 0, len(vaultPermissionValues))
	for name := range vaultPermissionValues {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}()

var vaultGroupPermissionAttrTypes = map[string]attr.Type{
	"group_id":    types.StringType,
	"permissions": types.SetType{ElemType: types.StringType},
}

type OnePasswordVaultResource struct {
	vaultManager onepassword.VaultManager
}

type OnePasswordVaultResourceModel struct {
	ID                types.String `tfsdk:"id"`
	UUID              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AllowAdminsAccess types.Bool   `tfsdk:"allow_admins_access"`
	GroupPermissions  types.Set    `tfsdk:"group_permissions"`
}

type OnePasswordVaultGroupPermissionModel struct {
	GroupID     types.String `tfsdk:"group_id"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func NewOnePasswordVaultResource() resource.Resource {
	return &OnePasswordVaultResource{}
}

func (r *OnePasswordVaultResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (r *OnePasswordVaultResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a 1Password vault using service account or desktop app authentication. Vault management is not available with 1Password Connect. Service accounts can only access vaults assigned to them or created by them, cannot manage vaults created using another authentication method, cannot update vault metadata, and can only delete or manage group permissions for vaults they created.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Terraform resource identifier in the format `vaults/<vault_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the vault.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the vault.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the vault.",
				Optional:            true,
			},
			"allow_admins_access": schema.BoolAttribute{
				MarkdownDescription: "Whether members of the Administrators group can access the vault. Defaults to `true`. This is only configurable when the vault is created, so changing it replaces the vault.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"group_permissions": schema.SetNestedBlock{
				MarkdownDescription: "Group access managed by this resource. Repeat this block to manage multiple groups. Groups omitted from this block are not managed, and importing a vault does not automatically adopt its existing group access. The 1Password SDK currently identifies groups by UUID and does not expose lookup by name.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the group.",
							Required:            true,
						},
						"permissions": schema.SetAttribute{
							MarkdownDescription: fmt.Sprintf("The complete permission set for the group. Supported values: `%s`. Permission dependencies vary by account type and are validated by 1Password.", strings.Join(vaultPermissionNames, "`, `")),
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(stringvalidator.OneOf(vaultPermissionNames...)),
							},
						},
					},
				},
			},
		},
	}
}

func (r *OnePasswordVaultResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	vaultManager, ok := client.(onepassword.VaultManager)
	if !ok {
		resp.Diagnostics.AddError(
			"Unsupported authentication for onepassword_vault",
			"The onepassword_vault resource cannot be used with 1Password Connect authentication. Use service account or desktop app authentication to manage vaults.",
		)
		return
	}
	r.vaultManager = vaultManager
}

func (r *OnePasswordVaultResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config OnePasswordVaultResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.GroupPermissions.IsNull() || config.GroupPermissions.IsUnknown() {
		return
	}

	var groups []OnePasswordVaultGroupPermissionModel
	resp.Diagnostics.Append(config.GroupPermissions.ElementsAs(ctx, &groups, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	seenGroupIDs := make(map[string]struct{}, len(groups))
	for i, group := range groups {
		if group.GroupID.IsUnknown() {
			continue
		}
		if _, exists := seenGroupIDs[group.GroupID.ValueString()]; exists {
			resp.Diagnostics.AddAttributeError(
				path.Root("group_permissions"),
				"Duplicate vault group",
				fmt.Sprintf("Group %q is configured more than once.", group.GroupID.ValueString()),
			)
		}
		seenGroupIDs[group.GroupID.ValueString()] = struct{}{}
		if group.Permissions.IsNull() || group.Permissions.IsUnknown() {
			continue
		}
		var names []string
		resp.Diagnostics.Append(group.Permissions.ElementsAs(ctx, &names, false)...)
		if len(names) > 1 && containsString(names, "no_access") {
			resp.Diagnostics.AddAttributeError(
				path.Root("group_permissions").AtSetValue(config.GroupPermissions.Elements()[i]).AtName("permissions"),
				"Invalid vault permissions",
				"`no_access` cannot be combined with other permissions.",
			)
		}
	}
}

func (r *OnePasswordVaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnePasswordVaultResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vault := &model.Vault{Name: plan.Name.ValueString(), Description: plan.Description.ValueString()}
	created, err := r.vaultManager.CreateVault(ctx, vault, plan.AllowAdminsAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("1Password Vault create error", fmt.Sprintf("Could not create vault: %s", err))
		return
	}

	plan.ID = types.StringValue(vaultTerraformID(created))
	plan.UUID = types.StringValue(created.ID)
	plan.Name = types.StringValue(created.Name)
	plan.Description = setStringValuePreservingEmpty(created.Description, plan.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	access, diagnostics := vaultAccessFromTerraform(ctx, plan.GroupPermissions)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(access) > 0 {
		if err := r.vaultManager.GrantVaultGroupPermissions(ctx, created.ID, access); err != nil {
			cleanupErr := r.vaultManager.DeleteVault(ctx, created.ID)
			if cleanupErr == nil {
				resp.State.RemoveResource(ctx)
				resp.Diagnostics.AddError("1Password Vault group permission error", fmt.Sprintf("Could not grant group permissions; the newly created vault was deleted: %s", err))
			} else {
				resp.Diagnostics.AddError("1Password Vault group permission error", fmt.Sprintf("Could not grant group permissions: %s. Cleanup also failed; vault %q may need to be deleted manually: %s", err, created.ID, cleanupErr))
			}
			return
		}
	}

	tflog.Trace(ctx, "created a vault", map[string]any{"vault_uuid": created.ID})
}

func (r *OnePasswordVaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnePasswordVaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := vaultUUIDFromTerraformID(state.ID.ValueString())
	if uuid == "" {
		resp.Diagnostics.AddError("Invalid 1Password Vault ID", fmt.Sprintf("Expected an ID in the format `vaults/<vault_id>`, got %q.", state.ID.ValueString()))
		return
	}

	vault, err := r.vaultManager.GetVaultDetails(ctx, uuid)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("1Password Vault read error", fmt.Sprintf("Could not get vault %q: %s", uuid, err))
		return
	}

	state.ID = types.StringValue(vaultTerraformID(vault))
	state.UUID = types.StringValue(vault.ID)
	state.Name = types.StringValue(vault.Name)
	state.Description = setStringValuePreservingEmpty(vault.Description, state.Description)
	refreshedAccess, diagnostics := refreshManagedVaultAccess(ctx, state.GroupPermissions, vault.GroupAccess)
	resp.Diagnostics.Append(diagnostics...)
	state.GroupPermissions = refreshedAccess
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnePasswordVaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnePasswordVaultResourceModel
	var state OnePasswordVaultResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := vaultUUIDFromTerraformID(state.ID.ValueString())
	if uuid == "" {
		resp.Diagnostics.AddError("Invalid 1Password Vault ID", fmt.Sprintf("Expected an ID in the format `vaults/<vault_id>`, got %q.", state.ID.ValueString()))
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {
		vault := &model.Vault{ID: uuid, Name: plan.Name.ValueString(), Description: plan.Description.ValueString()}
		updated, err := r.vaultManager.UpdateVault(ctx, vault)
		if err != nil {
			resp.Diagnostics.AddError("1Password Vault update error", fmt.Sprintf("Could not update vault %q: %s", uuid, err))
			return
		}
		plan.Name = types.StringValue(updated.Name)
		plan.Description = setStringValuePreservingEmpty(updated.Description, plan.Description)
	}

	previous, diagnostics := vaultAccessFromTerraform(ctx, state.GroupPermissions)
	resp.Diagnostics.Append(diagnostics...)
	desired, diagnostics := vaultAccessFromTerraform(ctx, plan.GroupPermissions)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.reconcileVaultGroupPermissions(ctx, uuid, previous, desired); err != nil {
		resp.Diagnostics.AddError("1Password Vault group permission error", fmt.Sprintf("Could not reconcile group permissions for vault %q: %s", uuid, err))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("vaults/%s", uuid))
	plan.UUID = types.StringValue(uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnePasswordVaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnePasswordVaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid := vaultUUIDFromTerraformID(state.ID.ValueString())
	if uuid == "" {
		resp.Diagnostics.AddError("Invalid 1Password Vault ID", fmt.Sprintf("Expected an ID in the format `vaults/<vault_id>`, got %q.", state.ID.ValueString()))
		return
	}
	if err := r.vaultManager.DeleteVault(ctx, uuid); err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("1Password Vault delete error", fmt.Sprintf("Could not delete vault %q: %s", uuid, err))
	}
}

func (r *OnePasswordVaultResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if vaultUUIDFromTerraformID(req.ID) == "" {
		resp.Diagnostics.AddError("Invalid 1Password Vault import ID", fmt.Sprintf("Expected an ID in the format `vaults/<vault_id>`, got %q.", req.ID))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *OnePasswordVaultResource) reconcileVaultGroupPermissions(ctx context.Context, uuid string, previous, desired []model.VaultGroupAccess) error {
	previousByGroup := vaultAccessByGroup(previous)
	desiredByGroup := vaultAccessByGroup(desired)
	var grants []model.VaultGroupAccess
	var updates []model.VaultGroupAccess

	for groupID, access := range desiredByGroup {
		old, exists := previousByGroup[groupID]
		if !exists {
			grants = append(grants, access)
		} else if old.Permissions != access.Permissions {
			updates = append(updates, access)
		}
	}
	if len(grants) > 0 {
		if err := r.vaultManager.GrantVaultGroupPermissions(ctx, uuid, grants); err != nil {
			return err
		}
	}
	if len(updates) > 0 {
		if err := r.vaultManager.UpdateVaultGroupPermissions(ctx, uuid, updates); err != nil {
			return err
		}
	}
	for groupID := range previousByGroup {
		if _, exists := desiredByGroup[groupID]; exists {
			continue
		}
		if err := r.vaultManager.RevokeVaultGroupPermission(ctx, uuid, groupID); err != nil {
			return err
		}
	}
	return nil
}

func vaultAccessFromTerraform(ctx context.Context, value types.Set) ([]model.VaultGroupAccess, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diagnostics
	}

	var groups []OnePasswordVaultGroupPermissionModel
	diagnostics.Append(value.ElementsAs(ctx, &groups, false)...)
	result := make([]model.VaultGroupAccess, 0, len(groups))
	for _, group := range groups {
		var permissionNames []string
		diagnostics.Append(group.Permissions.ElementsAs(ctx, &permissionNames, false)...)
		var bitmask uint32
		for _, name := range permissionNames {
			permission, exists := vaultPermissionValues[name]
			if !exists {
				diagnostics.AddError("Invalid vault permission", fmt.Sprintf("Unsupported permission %q.", name))
				continue
			}
			bitmask |= permission
		}
		result = append(result, model.VaultGroupAccess{
			GroupID:     group.GroupID.ValueString(),
			Permissions: bitmask,
		})
	}
	return result, diagnostics
}

func refreshManagedVaultAccess(ctx context.Context, managed types.Set, actual []model.VaultGroupAccess) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if managed.IsNull() || managed.IsUnknown() {
		return managed, diagnostics
	}

	var configured []OnePasswordVaultGroupPermissionModel
	diagnostics.Append(managed.ElementsAs(ctx, &configured, false)...)
	if diagnostics.HasError() {
		return managed, diagnostics
	}

	actualByGroup := vaultAccessByGroup(actual)
	refreshed := make([]OnePasswordVaultGroupPermissionModel, 0, len(configured))
	for _, configuredGroup := range configured {
		access, exists := actualByGroup[configuredGroup.GroupID.ValueString()]
		if !exists {
			continue
		}
		permissions, permissionDiagnostics := terraformPermissionsFromBitmask(ctx, access.Permissions)
		diagnostics.Append(permissionDiagnostics...)
		refreshed = append(refreshed, OnePasswordVaultGroupPermissionModel{
			GroupID:     configuredGroup.GroupID,
			Permissions: permissions,
		})
	}

	result, setDiagnostics := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: vaultGroupPermissionAttrTypes}, refreshed)
	diagnostics.Append(setDiagnostics...)
	return result, diagnostics
}

func terraformPermissionsFromBitmask(ctx context.Context, bitmask uint32) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if bitmask == model.VaultPermissionNoAccess {
		result, setDiagnostics := types.SetValueFrom(ctx, types.StringType, []string{"no_access"})
		diagnostics.Append(setDiagnostics...)
		return result, diagnostics
	}

	remaining := bitmask
	names := make([]string, 0, len(vaultPermissionValues))
	for _, name := range vaultPermissionNames {
		permission := vaultPermissionValues[name]
		if permission == model.VaultPermissionNoAccess {
			continue
		}
		if bitmask&permission != 0 {
			names = append(names, name)
			remaining &^= permission
		}
	}
	if remaining != 0 {
		diagnostics.AddError("Unsupported vault permission bitmask", fmt.Sprintf("1Password returned unknown permission bits %d.", remaining))
	}
	result, setDiagnostics := types.SetValueFrom(ctx, types.StringType, names)
	diagnostics.Append(setDiagnostics...)
	return result, diagnostics
}

func vaultUUIDFromTerraformID(id string) string {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] != "vaults" || parts[1] == "" {
		return ""
	}
	return parts[1]
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func vaultAccessByGroup(access []model.VaultGroupAccess) map[string]model.VaultGroupAccess {
	result := make(map[string]model.VaultGroupAccess, len(access))
	for _, permission := range access {
		result[permission.GroupID] = permission
	}
	return result
}
