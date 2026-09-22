package provider

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func ValueModifier() planmodifier.String {
	return valueModifier{}
}

type valueModifier struct{}

func (m valueModifier) Description(_ context.Context) string {
	// TODO: Come up with a better description
	return "Once set, the value of this attribute in state will not change unless the password recipe is changed."
}

func (m valueModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change unless the password recipe is changed."
}

func (m valueModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Check if the password recipe is changed. If so, then the value will be recomputed.
	var statePasswordRecipe, planPasswordRecipe []PasswordRecipeModel

	passwordRecipePath := req.Path.ParentPath().AtName("password_recipe")

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, passwordRecipePath, &statePasswordRecipe)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, passwordRecipePath, &planPasswordRecipe)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if shouldPreservePasswordValue(statePasswordRecipe, planPasswordRecipe) {
		resp.PlanValue = req.StateValue
	}
}

// shouldPreservePasswordValue returns true when the modifier should copy state password to plan
func shouldPreservePasswordValue(stateRecipe, planRecipe []PasswordRecipeModel) bool {
	if reflect.DeepEqual(stateRecipe, planRecipe) {
		return true
	}

	stateHasNoRecipe := len(stateRecipe) == 0
	planHasRecipe := len(planRecipe) > 0

	// If the state has no recipe and the plan has a recipe then the password should be preserved.
	if stateHasNoRecipe && planHasRecipe {
		return true
	}

	// Otherwise, the password should not be preserved.
	return false
}

// PasswordValueModifierForMapField is used for map-based fields where password_recipe is a SingleNestedAttribute
func PasswordValueModifierForMapField() planmodifier.String {
	return passwordValueModifierForMapField{}
}

type passwordValueModifierForMapField struct{}

func (m passwordValueModifierForMapField) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change unless the password recipe is changed."
}

func (m passwordValueModifierForMapField) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change unless the password recipe is changed."
}

func (m passwordValueModifierForMapField) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Check if the password recipe is changed. If so, then the value will be recomputed.
	var statePasswordRecipe, planPasswordRecipe *PasswordRecipeModel

	passwordRecipePath := req.Path.ParentPath().AtName("password_recipe")

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, passwordRecipePath, &statePasswordRecipe)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, passwordRecipePath, &planPasswordRecipe)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !reflect.DeepEqual(statePasswordRecipe, planPasswordRecipe) {
		return
	}

	resp.PlanValue = req.StateValue
}
