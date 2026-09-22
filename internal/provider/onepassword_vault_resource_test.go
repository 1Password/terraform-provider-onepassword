package provider

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

func TestAccVaultResourceRejectsConnect(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccProviderConfig("http://127.0.0.1:1") + `
resource "onepassword_vault" "test" {
  name = "test-vault"
}`,
			ExpectError: regexp.MustCompile(`onepassword_vault resource cannot be used`),
		}},
	})
}

func TestVaultPermissionRoundTrip(t *testing.T) {
	ctx := context.Background()
	permissions, diagnostics := terraformPermissionsFromBitmask(
		ctx,
		model.VaultPermissionReadItems|model.VaultPermissionCreateItems|model.VaultPermissionManageVault,
	)
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}

	groups, groupDiagnostics := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: vaultGroupPermissionAttrTypes}, []OnePasswordVaultGroupPermissionModel{{
		GroupID:     types.StringValue("group1"),
		Permissions: permissions,
	}})
	if groupDiagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", groupDiagnostics)
	}

	access, diagnostics := vaultAccessFromTerraform(ctx, groups)
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}
	if len(access) != 1 {
		t.Fatalf("Expected one group, got %d", len(access))
	}
	expected := model.VaultPermissionReadItems | model.VaultPermissionCreateItems | model.VaultPermissionManageVault
	if access[0].GroupID != "group1" || access[0].Permissions != expected {
		t.Fatalf("Unexpected access: %+v", access[0])
	}
}

func TestRefreshManagedVaultAccess(t *testing.T) {
	ctx := context.Background()
	permissions, diagnostics := types.SetValueFrom(ctx, types.StringType, []string{"read_items"})
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}
	managed, diagnostics := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: vaultGroupPermissionAttrTypes}, []OnePasswordVaultGroupPermissionModel{
		{GroupID: types.StringValue("managed"), Permissions: permissions},
		{GroupID: types.StringValue("removed"), Permissions: permissions},
	})
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}

	refreshed, diagnostics := refreshManagedVaultAccess(ctx, managed, []model.VaultGroupAccess{
		{GroupID: "managed", Permissions: model.VaultPermissionReadItems | model.VaultPermissionCreateItems},
		{GroupID: "external", Permissions: model.VaultPermissionManageVault},
	})
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}

	var groups []OnePasswordVaultGroupPermissionModel
	diagnostics = refreshed.ElementsAs(ctx, &groups, false)
	if diagnostics.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diagnostics)
	}
	if len(groups) != 1 || groups[0].GroupID.ValueString() != "managed" {
		t.Fatalf("Expected only the configured, existing group, got %+v", groups)
	}
}

func TestVaultUUIDFromTerraformID(t *testing.T) {
	tests := map[string]string{
		"vaults/abc":       "abc",
		"abc":              "",
		"vaults/":          "",
		"items/abc":        "",
		"vaults/abc/extra": "",
	}
	for input, expected := range tests {
		if actual := vaultUUIDFromTerraformID(input); actual != expected {
			t.Errorf("vaultUUIDFromTerraformID(%q) = %q, expected %q", input, actual, expected)
		}
	}
}
