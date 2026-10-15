package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestUseGeneratedIDForUnknown_EmptyState guards the actual bug this
// modifier fixes: a section/field id recorded in state as an empty string
// (not null) must not be locked in as the planned value, because
// toModelSectionField/toModelSectionFieldMap/toModelSections always replace
// an empty id with a freshly generated one on Create/Update. Locking "" in as
// the plan value makes Terraform reject the apply with "Provider produced
// inconsistent result after apply" the moment a real id gets assigned.
func TestUseGeneratedIDForUnknown_EmptyState(t *testing.T) {
	cases := []struct {
		name        string
		stateValue  types.String
		planValue   types.String
		configValue types.String
		wantPlan    types.String
	}{
		{
			name:        "empty string state is treated like null - plan stays unknown",
			stateValue:  types.StringValue(""),
			planValue:   types.StringUnknown(),
			configValue: types.StringNull(),
			wantPlan:    types.StringUnknown(),
		},
		{
			name:        "null state - plan stays unknown",
			stateValue:  types.StringNull(),
			planValue:   types.StringUnknown(),
			configValue: types.StringNull(),
			wantPlan:    types.StringUnknown(),
		},
		{
			name:        "non-empty state is locked in as the planned value",
			stateValue:  types.StringValue("f3a8d774-47c7-708b-c997-b7b65d44b344"),
			planValue:   types.StringUnknown(),
			configValue: types.StringNull(),
			wantPlan:    types.StringValue("f3a8d774-47c7-708b-c997-b7b65d44b344"),
		},
		{
			name:        "already-known plan value is left untouched",
			stateValue:  types.StringValue(""),
			planValue:   types.StringValue("something-else"),
			configValue: types.StringNull(),
			wantPlan:    types.StringValue("something-else"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue:  tc.stateValue,
				PlanValue:   tc.planValue,
				ConfigValue: tc.configValue,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tc.planValue,
			}

			UseGeneratedIDForUnknown().PlanModifyString(context.Background(), req, resp)

			if !resp.PlanValue.Equal(tc.wantPlan) {
				t.Errorf("PlanValue = %#v, want %#v", resp.PlanValue, tc.wantPlan)
			}
		})
	}
}
