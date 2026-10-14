package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateDate() dateValidator {
	return dateValidator{}
}

type dateValidator struct{}

func (v dateValidator) Description(_ context.Context) string {
	return "DATE values must be in YYYY-MM-DD format"
}

func (v dateValidator) MarkdownDescription(_ context.Context) string {
	return "DATE values must be in YYYY-MM-DD format"
}

func (v dateValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()

	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid DATE format",
			fmt.Sprintf("DATE values must be in YYYY-MM-DD format (e.g., 2026-09-20), got: %s", value),
		)
	}
}
