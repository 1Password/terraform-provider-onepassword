package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateType() validator.String {
	return typeValidator{}
}

type typeValidator struct{}

func (v typeValidator) Description(_ context.Context) string {
	return "Allowed type values are dependent on the item category."
}

func (v typeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v typeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var category types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("category"), &category)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if category.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()

	var allowed []string
	switch strings.ToLower(category.ValueString()) {
	case "database":
		allowed = dbTypes
	case "api_credential":
		allowed = apiCredentialTypes
	default:
		// `type` is not meaningful for other categories; skip.
		return
	}

	for _, a := range allowed {
		if strings.EqualFold(a, val) {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		fmt.Sprintf("Invalid type for category %q", category),
		fmt.Sprintf("Type for category %q must be one of %v, got: %q", category, allowed, val),
	)
}
