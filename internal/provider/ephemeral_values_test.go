package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

// sectionMapItem builds a model with one section_map field. version and woValue are what
// configuration sets; the plan carries the field ID from state and an unknown value.
func sectionMapItem(fieldID string, version types.Int64, woValue types.String, value types.String) *OnePasswordItemResourceModel {
	return &OnePasswordItemResourceModel{
		SectionMap: map[string]OnePasswordItemResourceSectionMapModel{
			"Credential": {
				FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
					"Token": {
						ID:             types.StringValue(fieldID),
						Value:          value,
						ValueWO:        woValue,
						ValueWOVersion: version,
					},
				},
			},
		},
	}
}

func sectionListItem(fieldID string, version types.Int64, woValue types.String, value types.String) *OnePasswordItemResourceModel {
	return &OnePasswordItemResourceModel{
		SectionList: []OnePasswordItemResourceSectionListModel{{
			Label: types.StringValue("Credential"),
			FieldList: []OnePasswordItemResourceFieldModel{{
				ID:             types.StringValue(fieldID),
				Label:          types.StringValue("Token"),
				Value:          value,
				ValueWO:        woValue,
				ValueWOVersion: version,
			}},
		}},
	}
}

func currentItemWithToken(value string) *model.Item {
	return &model.Item{
		Sections: []model.ItemSection{{ID: "s1", Label: "Credential"}},
		Fields: []model.ItemField{{
			ID:        "f1",
			SectionID: "s1",
			Label:     "Token",
			Value:     value,
		}},
	}
}

func mapPlanValue(m *OnePasswordItemResourceModel) types.String {
	return m.SectionMap["Credential"].FieldMap["Token"].Value
}

func listPlanValue(m *OnePasswordItemResourceModel) types.String {
	return m.SectionList[0].FieldList[0].Value
}

func TestApplySectionFieldWriteOnlyValues(t *testing.T) {
	type build func(string, types.Int64, types.String, types.String) *OnePasswordItemResourceModel
	type read func(*OnePasswordItemResourceModel) types.String

	forms := map[string]struct {
		build build
		value read
	}{
		"section_map": {sectionMapItem, mapPlanValue},
		"section":     {sectionListItem, listPlanValue},
	}

	tests := map[string]struct {
		stateVersion  *int64
		configVersion int64
		fieldID       string
		wantValue     string
		wantFetch     bool
	}{
		"create writes value_wo": {
			configVersion: 1,
			fieldID:       "",
			wantValue:     "new-secret",
		},
		"increased version writes value_wo": {
			stateVersion:  ptr(int64(1)),
			configVersion: 2,
			fieldID:       "f1",
			wantValue:     "new-secret",
		},
		"field newly switched to value_wo writes it": {
			stateVersion:  nil,
			configVersion: 1,
			fieldID:       "f1",
			wantValue:     "new-secret",
		},
		"same version keeps the value in 1Password": {
			stateVersion:  ptr(int64(1)),
			configVersion: 1,
			fieldID:       "f1",
			wantValue:     "current-secret",
			wantFetch:     true,
		},
		"decreased version keeps the value in 1Password": {
			stateVersion:  ptr(int64(3)),
			configVersion: 2,
			fieldID:       "f1",
			wantValue:     "current-secret",
			wantFetch:     true,
		},
		"field without an ID is found by label": {
			stateVersion:  ptr(int64(1)),
			configVersion: 1,
			fieldID:       "",
			wantValue:     "current-secret",
			wantFetch:     true,
		},
	}

	for formName, form := range forms {
		for name, tt := range tests {
			t.Run(formName+"/"+name, func(t *testing.T) {
				config := form.build("", types.Int64Value(tt.configVersion), types.StringValue("new-secret"), types.StringNull())
				plan := form.build(tt.fieldID, types.Int64Value(tt.configVersion), types.StringNull(), types.StringUnknown())

				var state *OnePasswordItemResourceModel
				if tt.stateVersion != nil {
					state = form.build("f1", types.Int64Value(*tt.stateVersion), types.StringNull(), types.StringNull())
				} else if tt.fieldID != "" {
					state = form.build("f1", types.Int64Null(), types.StringNull(), types.StringValue("old-plain-value"))
				}

				fetched := false
				err := applySectionFieldWriteOnlyValues(config, state, plan, func() (*model.Item, error) {
					fetched = true
					return currentItemWithToken("current-secret"), nil
				})
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				if got := form.value(plan); got.ValueString() != tt.wantValue {
					t.Errorf("plan value = %q, want %q", got.ValueString(), tt.wantValue)
				}
				if fetched != tt.wantFetch {
					t.Errorf("fetched current item = %v, want %v", fetched, tt.wantFetch)
				}
			})
		}
	}
}

func TestApplySectionFieldWriteOnlyValues_FieldMissingIn1Password(t *testing.T) {
	config := sectionMapItem("", types.Int64Value(1), types.StringValue("new-secret"), types.StringNull())
	state := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringNull())
	plan := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringUnknown())

	err := applySectionFieldWriteOnlyValues(config, state, plan, func() (*model.Item, error) {
		return &model.Item{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := mapPlanValue(plan); !got.IsNull() {
		t.Errorf("plan value = %v, want null", got)
	}
}

func TestApplySectionFieldWriteOnlyValues_ReadError(t *testing.T) {
	config := sectionMapItem("", types.Int64Value(1), types.StringValue("new-secret"), types.StringNull())
	state := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringNull())
	plan := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringUnknown())

	err := applySectionFieldWriteOnlyValues(config, state, plan, func() (*model.Item, error) {
		return nil, errors.New("read failed")
	})
	if err == nil {
		t.Fatal("expected the read error to be returned")
	}
}

// Under ignore_changes the plan keeps the prior state's fields. A field that was not write-only
// in state keeps its value even though configuration now sets value_wo.
func TestApplySectionFieldWriteOnlyValues_IgnoredFieldNotYetWriteOnly(t *testing.T) {
	config := sectionMapItem("", types.Int64Value(1), types.StringValue(""), types.StringNull())
	state := sectionMapItem("f1", types.Int64Null(), types.StringNull(), types.StringValue("hand-entered"))
	plan := sectionMapItem("f1", types.Int64Null(), types.StringNull(), types.StringValue("hand-entered"))

	fetched := false
	err := applySectionFieldWriteOnlyValues(config, state, plan, func() (*model.Item, error) {
		fetched = true
		return currentItemWithToken("current-secret"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := mapPlanValue(plan); got.ValueString() != "hand-entered" {
		t.Errorf("plan value = %q, want the planned value kept", got.ValueString())
	}
	if fetched {
		t.Error("fetched the current item for a field the plan does not mark write-only")
	}
}

// Under ignore_changes a write-only field keeps the value in 1Password, even where configuration
// no longer has the field or has raised its version.
func TestApplySectionFieldWriteOnlyValues_IgnoredWriteOnlyField(t *testing.T) {
	configs := map[string]*OnePasswordItemResourceModel{
		"field removed from configuration": {},
		"version raised in configuration":  sectionMapItem("", types.Int64Value(2), types.StringValue("new-secret"), types.StringNull()),
	}

	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			state := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringNull())
			plan := sectionMapItem("f1", types.Int64Value(1), types.StringNull(), types.StringNull())

			err := applySectionFieldWriteOnlyValues(config, state, plan, func() (*model.Item, error) {
				return currentItemWithToken("hand-entered"), nil
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got := mapPlanValue(plan); got.ValueString() != "hand-entered" {
				t.Errorf("plan value = %q, want %q", got.ValueString(), "hand-entered")
			}
		})
	}
}

func TestClearSectionFieldWriteOnlyValues(t *testing.T) {
	m := &OnePasswordItemResourceModel{
		SectionMap: map[string]OnePasswordItemResourceSectionMapModel{
			"Credential": {
				FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
					"Token": {Value: types.StringValue("secret"), ValueWOVersion: types.Int64Value(1)},
					"ID":    {Value: types.StringValue("visible")},
				},
			},
		},
		SectionList: []OnePasswordItemResourceSectionListModel{{
			FieldList: []OnePasswordItemResourceFieldModel{
				{Value: types.StringValue("secret"), ValueWOVersion: types.Int64Value(1)},
				{Value: types.StringValue("visible")},
			},
		}},
	}

	clearSectionFieldWriteOnlyValues(m)

	fields := m.SectionMap["Credential"].FieldMap
	if !fields["Token"].Value.IsNull() {
		t.Errorf("section_map write-only value = %v, want null", fields["Token"].Value)
	}
	if fields["ID"].Value.ValueString() != "visible" {
		t.Errorf("section_map plain value = %v, want unchanged", fields["ID"].Value)
	}
	if !m.SectionList[0].FieldList[0].Value.IsNull() {
		t.Errorf("section write-only value = %v, want null", m.SectionList[0].FieldList[0].Value)
	}
	if m.SectionList[0].FieldList[1].Value.ValueString() != "visible" {
		t.Errorf("section plain value = %v, want unchanged", m.SectionList[0].FieldList[1].Value)
	}
}

func TestToStateSectionsAndFieldsMap_KeepsWriteOnlyVersion(t *testing.T) {
	item := currentItemWithToken("secret")
	prior := sectionMapItem("f1", types.Int64Value(4), types.StringNull(), types.StringNull())

	got := toStateSectionsAndFieldsMap(item, prior.SectionMap)

	if v := got["Credential"].FieldMap["Token"].ValueWOVersion; v.ValueInt64() != 4 {
		t.Errorf("value_wo_version = %v, want 4", v)
	}
}

func ptr[T any](v T) *T {
	return &v
}
