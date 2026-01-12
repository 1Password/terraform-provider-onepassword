package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var monthYearRegex = regexp.MustCompile(`^\d{4}(0[1-9]|1[0-2])$`)

func validateMonthYear() monthYearValidator {
	return monthYearValidator{}
}

type monthYearValidator struct{}

func (v monthYearValidator) Description(ctx context.Context) string {
	return "MONTH_YEAR values must be in YYYYMM format (e.g., 202401)"
}

func (v monthYearValidator) MarkdownDescription(ctx context.Context) string {
	return "MONTH_YEAR values must be in YYYYMM format (e.g., 202401)"
}

func (v monthYearValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var fieldType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &fieldType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if fieldType.ValueString() != "MONTH_YEAR" {
		return
	}

	value := req.ConfigValue.ValueString()

	if !monthYearRegex.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid MONTH_YEAR format",
			fmt.Sprintf("MONTH_YEAR values must be in YYYYMM format with valid month 01-12 (e.g., 202401), got: %s", value),
		)
	}
}
