package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func itemResourceSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&OnePasswordItemResource{}).Schema(ctx, resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	return resp.Schema
}

func nullItemRaw(ctx context.Context, s schema.Schema) tftypes.Value {
	return tftypes.NewValue(s.Type().TerraformType(ctx), nil)
}

func itemRaw(ctx context.Context, t *testing.T, s schema.Schema, item OnePasswordItemResourceModel) tftypes.Value {
	t.Helper()

	state := tfsdk.State{Schema: s, Raw: nullItemRaw(ctx, s)}
	if diags := state.Set(ctx, item); diags.HasError() {
		t.Fatalf("unexpected diagnostics setting state: %v", diags)
	}

	return state.Raw
}

func TestItemResourceModifyPlan(t *testing.T) {
	ctx := context.Background()
	s := itemResourceSchema(ctx, t)

	stored := OnePasswordItemResourceModel{
		ID:        types.StringValue("vaults/vault1/items/item1"),
		UUID:      types.StringValue("item1"),
		Vault:     types.StringValue("vault1"),
		Category:  types.StringValue("login"),
		Title:     types.StringValue("Test Item"),
		Tags:      types.ListNull(types.StringType),
		CreatedAt: types.StringValue("2024-03-01T10:30:00Z"),
		UpdatedAt: types.StringValue("2024-05-02T11:45:00Z"),
	}

	changed := stored
	changed.Title = types.StringValue("New Title")

	tests := map[string]struct {
		state         tftypes.Value
		plan          tftypes.Value
		expectUnknown bool
	}{
		"should leave updated_at untouched when nothing changed": {
			state:         itemRaw(ctx, t, s, stored),
			plan:          itemRaw(ctx, t, s, stored),
			expectUnknown: false,
		},
		"should mark updated_at unknown when a change is planned": {
			state:         itemRaw(ctx, t, s, stored),
			plan:          itemRaw(ctx, t, s, changed),
			expectUnknown: true,
		},
		"should do nothing on create": {
			state:         nullItemRaw(ctx, s),
			plan:          itemRaw(ctx, t, s, stored),
			expectUnknown: false,
		},
		"should do nothing on destroy": {
			state:         itemRaw(ctx, t, s, stored),
			plan:          nullItemRaw(ctx, s),
			expectUnknown: false,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			req := resource.ModifyPlanRequest{
				State: tfsdk.State{Schema: s, Raw: test.state},
				Plan:  tfsdk.Plan{Schema: s, Raw: test.plan},
			}
			resp := &resource.ModifyPlanResponse{
				Plan: tfsdk.Plan{Schema: s, Raw: test.plan},
			}

			(&OnePasswordItemResource{}).ModifyPlan(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			if resp.Plan.Raw.IsNull() {
				if test.expectUnknown {
					t.Fatal("expected updated_at to be unknown, but the plan is null")
				}
				return
			}

			var updatedAt types.String
			if diags := resp.Plan.GetAttribute(ctx, path.Root("updated_at"), &updatedAt); diags.HasError() {
				t.Fatalf("unexpected diagnostics reading updated_at: %v", diags)
			}

			if updatedAt.IsUnknown() != test.expectUnknown {
				t.Errorf("updated_at unknown = %v, expected %v (value %v)", updatedAt.IsUnknown(), test.expectUnknown, updatedAt)
			}
		})
	}
}
