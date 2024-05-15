package provider

import (
	"context"
	"fmt"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateOTP() otpValidator {
	return otpValidator{}
}

type otpValidator struct{}

func (v otpValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Fields of type OTP must have an ID with the '%s' prefix", OTPFieldIDPrefix)
}

func (v otpValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Fields of type OTP must have an ID with the '%s' prefix", OTPFieldIDPrefix)
}

func (v otpValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var fieldType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &fieldType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if fieldType.ValueString() == string(op.FieldTypeOTP) && !strings.HasPrefix(req.ConfigValue.ValueString(), OTPFieldIDPrefix) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ID for OTP type field",
			fmt.Sprintf("Field of type OTP must have the '%s' prefix, got: %s", OTPFieldIDPrefix, req.ConfigValue.ValueString()))
	}
}
