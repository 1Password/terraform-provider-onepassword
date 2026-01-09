package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToStateTags(t *testing.T) {
	ctx := context.Background()

	// Build test cases with context available
	testCases := []struct {
		name          string
		modelTags     []string
		stateTags     []string
		wantTags      []string
		stateTagsNull bool
		wantNull      bool
		wantErr       bool
	}{
		{
			name:      "tags match",
			modelTags: []string{"tag1", "tag2"},
			stateTags: []string{"tag1", "tag2"},
			wantTags:  []string{"tag1", "tag2"},
		},
		{
			name:      "single tag",
			modelTags: []string{"tag1"},
			stateTags: []string{"tag1"},
			wantTags:  []string{"tag1"},
		},
		{
			name:      "tags differ",
			modelTags: []string{"tag1", "tag2"},
			stateTags: []string{"tag3"},
			wantTags:  []string{"tag1", "tag2"},
		},
		{
			name:          "empty tags preserve null",
			modelTags:     []string{},
			stateTagsNull: true,
			wantNull:      true,
		},
		{
			name:      "empty tags with existing list",
			modelTags: []string{},
			stateTags: []string{"tag1"},
			wantTags:  []string{},
		},
		{
			name:          "null current tags",
			modelTags:     []string{"tag1"},
			stateTagsNull: true,
			wantTags:      []string{"tag1"},
		},
		{
			name:      "tags match with different order",
			modelTags: []string{"tag2", "tag1", "tag3"},
			stateTags: []string{"tag1", "tag2", "tag3"},
			wantTags:  []string{"tag1", "tag2", "tag3"},
		},
		{
			name:      "tags with special characters",
			modelTags: []string{"tag-1", "tag_2", "tag.3", "tag@4"},
			stateTags: []string{"tag-1", "tag_2", "tag.3", "tag@4"},
			wantTags:  []string{"tag-1", "tag_2", "tag.3", "tag@4"},
		},
		{
			name:      "empty string in tags",
			modelTags: []string{"tag1", "", "tag2"},
			stateTags: []string{"tag1", "tag2"},
			wantTags:  []string{"", "tag1", "tag2"},
		},
		{
			name:      "state has more tags than model",
			modelTags: []string{"tag1", "tag2"},
			stateTags: []string{"tag1", "tag2", "tag3", "tag4"},
			wantTags:  []string{"tag1", "tag2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert []string to types.List using context
			var stateTags, want types.List
			if tc.stateTagsNull {
				stateTags = types.ListNull(types.StringType)
			} else {
				stateTags, _ = types.ListValueFrom(ctx, types.StringType, tc.stateTags)
			}

			if tc.wantNull {
				want = types.ListNull(types.StringType)
			} else {
				want, _ = types.ListValueFrom(ctx, types.StringType, tc.wantTags)
			}

			got, diags := toStateTags(ctx, tc.modelTags, stateTags)
			if (diags.HasError()) != tc.wantErr {
				t.Errorf("processTags() error = %v, wantErr %v", diags.HasError(), tc.wantErr)
				return
			}
			if !tc.wantErr {
				if got.IsNull() != want.IsNull() {
					t.Errorf("processTags() IsNull = %v, want %v", got.IsNull(), want.IsNull())
					return
				}
				if !got.IsNull() {
					var gotSlice, wantSlice []string
					got.ElementsAs(ctx, &gotSlice, false)
					want.ElementsAs(ctx, &wantSlice, false)
					if !reflect.DeepEqual(gotSlice, wantSlice) {
						t.Errorf("processTags() = %v, want %v", gotSlice, wantSlice)
					}
				}
			}
		})
	}
}

func TestToStateSectionsAndFields(t *testing.T) {
	tests := []struct {
		name          string
		modelSections []model.ItemSection
		modelFields   []model.ItemField
		stateSections []OnePasswordItemResourceSectionListModel
		want          []OnePasswordItemResourceSectionListModel
	}{
		{
			name: "new section with field",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1")},
					},
				},
			},
		},
		{
			name: "update existing section by ID",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Updated Label"},
			},
			modelFields: []model.ItemField{},
			stateSections: []OnePasswordItemResourceSectionListModel{
				{
					ID:        types.StringValue("section1"),
					Label:     types.StringValue("Old Label"),
					FieldList: []OnePasswordItemResourceFieldModel{},
				},
			},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:        types.StringValue("section1"),
					Label:     types.StringValue("Updated Label"),
					FieldList: []OnePasswordItemResourceFieldModel{},
				},
			},
		},
		{
			name: "field with recipe",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{
					ID:        "field1",
					Label:     "Password",
					Type:      model.FieldTypeConcealed,
					SectionID: "section1",
					Recipe: &model.GeneratorRecipe{
						Length:        20,
						CharacterSets: []model.CharacterSet{model.CharacterSetDigits, model.CharacterSetSymbols},
					},
				},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Password"),
							Type:  types.StringValue("CONCEALED"),
							Recipe: []PasswordRecipeModel{
								{
									Length:  types.Int64Value(20),
									Digits:  types.BoolValue(true),
									Symbols: types.BoolValue(true),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "match section by label when ID is empty",
			modelSections: []model.ItemSection{
				{ID: "", Label: "Section 1"},
			},
			modelFields: []model.ItemField{},
			stateSections: []OnePasswordItemResourceSectionListModel{
				{
					ID:        types.StringNull(),
					Label:     types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{},
				},
			},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:        types.StringNull(),
					Label:     types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{},
				},
			},
		},
		{
			name: "match field by label when ID is empty",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{ID: "", Label: "Field 1", Type: model.FieldTypeString, Value: "updated value", SectionID: "section1"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringNull(),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("old value"),
						},
					},
				},
			},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringNull(),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("updated value"),
						},
					},
				},
			},
		},
		{
			name: "update existing field by ID",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "updated value", SectionID: "section1"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("old value"),
						},
					},
				},
			},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("updated value"),
						},
					},
				},
			},
		},
		{
			name: "multiple sections with multiple fields",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			modelFields: []model.ItemField{
				{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
				{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "section1"},
				{ID: "field3", Label: "Field 3", Type: model.FieldTypeString, Value: "value3", SectionID: "section2"},
				{ID: "field4", Label: "Field 4", Type: model.FieldTypeString, Value: "value4", SectionID: "section2"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
						},
						{
							ID:    types.StringValue("field2"),
							Label: types.StringValue("Field 2"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value2"),
						},
					},
				},
				{
					ID:    types.StringValue("section2"),
					Label: types.StringValue("Section 2"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field3"),
							Label: types.StringValue("Field 3"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value3"),
						},
						{
							ID:    types.StringValue("field4"),
							Label: types.StringValue("Field 4"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value4"),
						},
					},
				},
			},
		},
		{
			name: "field with SectionID that doesn't match any section is ignored",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
				{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "nonexistent-section"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Field 1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
						},
						// field2 should be ignored because its SectionID doesn't match any section
					},
				},
			},
		},
		{
			name: "section with empty label",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: ""},
			},
			modelFields:   []model.ItemField{},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:        types.StringValue("section1"),
					Label:     types.StringNull(),
					FieldList: []OnePasswordItemResourceFieldModel{},
				},
			},
		},
		{
			name: "field with empty label",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{ID: "field1", Label: "", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringNull(),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
						},
					},
				},
			},
		},
		{
			name: "field recipe with no character sets",
			modelSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			modelFields: []model.ItemField{
				{
					ID:        "field1",
					Label:     "Password",
					Type:      model.FieldTypeConcealed,
					SectionID: "section1",
					Recipe: &model.GeneratorRecipe{
						Length:        20,
						CharacterSets: []model.CharacterSet{},
					},
				},
			},
			stateSections: []OnePasswordItemResourceSectionListModel{},
			want: []OnePasswordItemResourceSectionListModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					FieldList: []OnePasswordItemResourceFieldModel{
						{
							ID:    types.StringValue("field1"),
							Label: types.StringValue("Password"),
							Type:  types.StringValue("CONCEALED"),
							Recipe: []PasswordRecipeModel{
								{
									Length:  types.Int64Value(20),
									Digits:  types.BoolValue(false),
									Symbols: types.BoolValue(false),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStateSectionsAndFieldsList(tt.modelSections, tt.modelFields, tt.stateSections)
			if len(got) != len(tt.want) {
				t.Errorf("processSectionsAndFields() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].ID.ValueString() != tt.want[i].ID.ValueString() {
					t.Errorf("Section[%d].ID = %v, want %v", i, got[i].ID.ValueString(), tt.want[i].ID.ValueString())
				}
				if got[i].Label.ValueString() != tt.want[i].Label.ValueString() {
					t.Errorf("Section[%d].Label = %v, want %v", i, got[i].Label.ValueString(), tt.want[i].Label.ValueString())
				}
				if len(got[i].FieldList) != len(tt.want[i].FieldList) {
					t.Errorf("Section[%d].FieldList len = %d, want %d", i, len(got[i].FieldList), len(tt.want[i].FieldList))
				}
			}
		})
	}
}

func TestToStateTopLevelFields(t *testing.T) {
	tests := []struct {
		name        string
		modelFields []model.ItemField
		state       *OnePasswordItemResourceModel
		want        *OnePasswordItemResourceModel
	}{
		{
			name: "username by purpose",
			modelFields: []model.ItemField{
				{Purpose: model.FieldPurposeUsername, Value: "user1"},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Username: types.StringValue("user1"),
			},
		},
		{
			name: "password by purpose",
			modelFields: []model.ItemField{
				{Purpose: model.FieldPurposePassword, Value: "pass1"},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Password: types.StringValue("pass1"),
			},
		},
		{
			name: "notes by purpose",
			modelFields: []model.ItemField{
				{Purpose: model.FieldPurposeNotes, Value: "note1"},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				NoteValue: types.StringValue("note1"),
			},
		},
		{
			name: "hostname by label",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Hostname: types.StringValue("example.com"),
			},
		}, {
			name: "port by label",
			modelFields: []model.ItemField{
				{Label: "port", Value: "3306", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Port: types.StringValue("3306"),
			},
		},
		{
			name: "type by label",
			modelFields: []model.ItemField{
				{Label: "type", Value: "mysql", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Type: types.StringValue("mysql"),
			},
		},
		{
			name: "field in section ignored",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: "section1"},
			},
			state: &OnePasswordItemResourceModel{},
			want:  &OnePasswordItemResourceModel{},
		},
		{
			name: "server label maps to hostname",
			modelFields: []model.ItemField{
				{Label: "server", Value: "server.example.com", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Hostname: types.StringValue("server.example.com"),
			},
		},
		{
			name: "username by label when purpose is empty",
			modelFields: []model.ItemField{
				{Label: "username", Value: "user1", SectionID: "", Purpose: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Username: types.StringValue("user1"),
			},
		},
		{
			name: "password by label when purpose is empty",
			modelFields: []model.ItemField{
				{Label: "password", Value: "pass1", SectionID: "", Purpose: ""},
			},
			state: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Password: types.StringValue("pass1"),
			},
		},
		{
			name: "existing values preserved when field not present in modelItem",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{
				Username: types.StringValue("existing-user"),
				Password: types.StringValue("existing-pass"),
				Database: types.StringValue("existing-db"),
				Port:     types.StringValue("5432"),
				Type:     types.StringValue("postgresql"),
			},
			want: &OnePasswordItemResourceModel{
				Username: types.StringValue("existing-user"), // Preserved
				Password: types.StringValue("existing-pass"), // Preserved
				Hostname: types.StringValue("example.com"),   // Updated
				Database: types.StringValue("existing-db"),   // Preserved
				Port:     types.StringValue("5432"),          // Preserved
				Type:     types.StringValue("postgresql"),    // Preserved
			},
		},
		{
			name:        "existing values preserved when modelFields is empty",
			modelFields: []model.ItemField{},
			state: &OnePasswordItemResourceModel{
				Username:  types.StringValue("existing-user"),
				Password:  types.StringValue("existing-pass"),
				Hostname:  types.StringValue("existing-host"),
				Database:  types.StringValue("existing-db"),
				Port:      types.StringValue("5432"),
				Type:      types.StringValue("postgresql"),
				NoteValue: types.StringValue("existing-note"),
			},
			want: &OnePasswordItemResourceModel{
				Username:  types.StringValue("existing-user"), // Preserved
				Password:  types.StringValue("existing-pass"), // Preserved
				Hostname:  types.StringValue("existing-host"), // Preserved
				Database:  types.StringValue("existing-db"),   // Preserved
				Port:      types.StringValue("5432"),          // Preserved
				Type:      types.StringValue("postgresql"),    // Preserved
				NoteValue: types.StringValue("existing-note"), // Preserved
			},
		},
		{
			name: "mix of fields present and absent preserves absent ones",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
				{Label: "port", Value: "3306", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{
				Username: types.StringValue("existing-user"),
				Password: types.StringValue("existing-pass"),
				Hostname: types.StringValue("old-host"),
				Database: types.StringValue("existing-db"),
				Port:     types.StringValue("5432"),
				Type:     types.StringValue("postgresql"),
			},
			want: &OnePasswordItemResourceModel{
				Username: types.StringValue("existing-user"), // Preserved
				Password: types.StringValue("existing-pass"), // Preserved
				Hostname: types.StringValue("example.com"),   // Updated
				Database: types.StringValue("existing-db"),   // Preserved
				Port:     types.StringValue("3306"),          // Updated
				Type:     types.StringValue("postgresql"),    // Preserved
			},
		},
		{
			name: "null state values remain null when field not present",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{
				Username: types.StringNull(),
				Password: types.StringNull(),
				Database: types.StringNull(),
				Port:     types.StringNull(),
				Type:     types.StringNull(),
			},
			want: &OnePasswordItemResourceModel{
				Username: types.StringNull(),               // Preserved as null
				Password: types.StringNull(),               // Preserved as null
				Hostname: types.StringValue("example.com"), // Updated
				Database: types.StringNull(),               // Preserved as null
				Port:     types.StringNull(),               // Preserved as null
				Type:     types.StringNull(),               // Preserved as null
			},
		},
		{
			name: "unknown state values preserved when field not present",
			modelFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
			},
			state: &OnePasswordItemResourceModel{
				Username: types.StringUnknown(),
				Password: types.StringUnknown(),
				Database: types.StringUnknown(),
			},
			want: &OnePasswordItemResourceModel{
				Username: types.StringUnknown(),            // Preserved as unknown
				Password: types.StringUnknown(),            // Preserved as unknown
				Hostname: types.StringValue("example.com"), // Updated
				Database: types.StringUnknown(),            // Preserved as unknown
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toStateTopLevelFields(tt.modelFields, tt.state)
			if tt.state.Username.ValueString() != tt.want.Username.ValueString() {
				t.Errorf("Username = %v, want %v", tt.state.Username.ValueString(), tt.want.Username.ValueString())
			}
			if tt.state.Password.ValueString() != tt.want.Password.ValueString() {
				t.Errorf("Password = %v, want %v", tt.state.Password.ValueString(), tt.want.Password.ValueString())
			}
			if tt.state.NoteValue.ValueString() != tt.want.NoteValue.ValueString() {
				t.Errorf("NoteValue = %v, want %v", tt.state.NoteValue.ValueString(), tt.want.NoteValue.ValueString())
			}
			if tt.state.Hostname.ValueString() != tt.want.Hostname.ValueString() {
				t.Errorf("Hostname = %v, want %v", tt.state.Hostname.ValueString(), tt.want.Hostname.ValueString())
			}
			if tt.state.Database.ValueString() != tt.want.Database.ValueString() {
				t.Errorf("Database = %v, want %v", tt.state.Database.ValueString(), tt.want.Database.ValueString())
			}
			if tt.state.Port.ValueString() != tt.want.Port.ValueString() {
				t.Errorf("Port = %v, want %v", tt.state.Port.ValueString(), tt.want.Port.ValueString())
			}
			if tt.state.Type.ValueString() != tt.want.Type.ValueString() {
				t.Errorf("Type = %v, want %v", tt.state.Type.ValueString(), tt.want.Type.ValueString())
			}
		})
	}
}

func TestToStateSectionsAndFieldsMap(t *testing.T) {
	tests := map[string]struct {
		item            *model.Item
		stateSectionMap map[string]OnePasswordItemResourceSectionMapModel
		want            map[string]OnePasswordItemResourceSectionMapModel
		wantErr         bool
	}{
		"new section with new field": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value1"),
							Recipe: nil,
						},
					},
				},
			},
		},
		"item from server only contains new field": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "section1"},
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						// Field 1 is removed - not in item from server
						"Field 2": {
							ID:     types.StringValue("field2"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value2"),
							Recipe: nil,
						},
					},
				},
			},
		},
		"existing section with existing field - update value": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "new-value", SectionID: "section1"},
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("old-value"),
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("new-value"), // Updated from model
							Recipe: nil,
						},
					},
				},
			},
		},
		"field with recipe - digits and symbols": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{
						ID:        "field1",
						Label:     "Password",
						Type:      model.FieldTypeConcealed,
						Value:     "Pass123!@#",
						SectionID: "section1",
						Recipe: &model.GeneratorRecipe{
							Length:        20,
							CharacterSets: []model.CharacterSet{model.CharacterSetDigits, model.CharacterSetSymbols},
						},
					},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringValue("Pass123!@#"),
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true), // Has digits
								Symbols: types.BoolValue(true), // Has symbols
							},
						},
					},
				},
			},
		},
		"field with recipe - no digits or symbols": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{
						ID:        "field1",
						Label:     "Password",
						Type:      model.FieldTypeConcealed,
						Value:     "PasswordOnly",
						SectionID: "section1",
						Recipe: &model.GeneratorRecipe{
							Length:        15,
							CharacterSets: []model.CharacterSet{},
						},
					},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringValue("PasswordOnly"),
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(15),
								Digits:  types.BoolValue(false), // No digits
								Symbols: types.BoolValue(false), // No symbols
							},
						},
					},
				},
			},
		},
		"field without recipe in model - preserves existing recipe from plan": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1", Recipe: nil},
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("value1"),
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
		},
		"field with recipe in state and empty server value - server value used, recipe preserved": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Password", Type: model.FieldTypeConcealed, Value: "", SectionID: "section1", Recipe: nil}, // Server returns empty
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringValue("state-password-123"), // Password from state - NOT preserved
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringNull(), // Server value (empty -> null), NOT state value
							Recipe: &PasswordRecipeModel{ // Recipe IS preserved from state
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
		},
		"field with recipe in state and different server value - server value used, recipe preserved": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					// Server returns different password (e.g., user changed it in 1Password directly)
					{ID: "field1", Label: "Password", Type: model.FieldTypeConcealed, Value: "server-password-456", SectionID: "section1", Recipe: nil},
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringValue("old-password-123"), // Password from state - NOT preserved
							Recipe: &PasswordRecipeModel{
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Password": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("CONCEALED"),
							Value: types.StringValue("server-password-456"), // Server is source of truth for VALUE
							Recipe: &PasswordRecipeModel{ // Recipe IS preserved from state
								Length:  types.Int64Value(20),
								Digits:  types.BoolValue(true),
								Symbols: types.BoolValue(true),
							},
						},
					},
				},
			},
		},
		"multiple sections with multiple fields": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
					{ID: "section2", Label: "Section 2"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
					{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "section1"},
					{ID: "field3", Label: "Field 3", Type: model.FieldTypeString, Value: "value3", SectionID: "section2"},
					{ID: "field4", Label: "Field 4", Type: model.FieldTypeString, Value: "value4", SectionID: "section2"},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value1"),
							Recipe: nil,
						},
						"Field 2": {
							ID:     types.StringValue("field2"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value2"),
							Recipe: nil,
						},
					},
				},
				"Section 2": {
					ID: types.StringValue("section2"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 3": {
							ID:     types.StringValue("field3"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value3"),
							Recipe: nil,
						},
						"Field 4": {
							ID:     types.StringValue("field4"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value4"),
							Recipe: nil,
						},
					},
				},
			},
		},
		"field that doesn't belong to section - ignored": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
					{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "nonexistent-section"},
					{ID: "field3", Label: "Field 3", Type: model.FieldTypeString, Value: "value3", SectionID: ""},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value1"),
							Recipe: nil,
						},
						// Field 2 and Field 3 should be ignored
					},
				},
			},
		},
		"empty state map": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
				},
			},
			stateSectionMap: nil,
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value1"),
							Recipe: nil,
						},
					},
				},
			},
		},
		"item from server represents complete state - fields not in item are removed": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
				},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:    types.StringValue("field1"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("old-value"),
						},
						"Field 2": { // This field is not in item from server, should be removed
							ID:    types.StringValue("field2"),
							Type:  types.StringValue("STRING"),
							Value: types.StringValue("preserved-value"),
						},
					},
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 1": {
							ID:     types.StringValue("field1"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value1"), // Updated from item
							Recipe: nil,
						},
						// Field 2 is removed - not in item from server
					},
				},
			},
		},
		"section ID updated from model": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "new-section-id", Label: "Section 1"},
				},
				Fields: []model.ItemField{},
			},
			stateSectionMap: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID:       types.StringValue("old-section-id"),
					FieldMap: make(map[string]OnePasswordItemResourceFieldMapModel),
				},
			},
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID:       types.StringValue("new-section-id"), // Updated from model (API response) - required for create/update
					FieldMap: make(map[string]OnePasswordItemResourceFieldMapModel),
				},
			},
		},
		"empty section with no fields": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID:       types.StringValue("section1"),
					FieldMap: make(map[string]OnePasswordItemResourceFieldMapModel),
				},
			},
		},
		"section with empty label - last wins when multiple sections have empty labels": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: ""},
					{ID: "section2", Label: ""},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
					{ID: "field2", Label: "Field 2", Type: model.FieldTypeString, Value: "value2", SectionID: "section2"},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"": {
					ID: types.StringValue("section2"), // Last section ID wins
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"Field 2": {
							ID:     types.StringValue("field2"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value2"),
							Recipe: nil,
						},
					},
				},
			},
		},
		"field with empty label - last wins when multiple fields have empty labels": {
			item: &model.Item{
				Sections: []model.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []model.ItemField{
					{ID: "field1", Label: "", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
					{ID: "field2", Label: "", Type: model.FieldTypeString, Value: "value2", SectionID: "section1"},
				},
			},
			stateSectionMap: make(map[string]OnePasswordItemResourceSectionMapModel),
			want: map[string]OnePasswordItemResourceSectionMapModel{
				"Section 1": {
					ID: types.StringValue("section1"),
					FieldMap: map[string]OnePasswordItemResourceFieldMapModel{
						"": {
							ID:     types.StringValue("field2"),
							Type:   types.StringValue("STRING"),
							Value:  types.StringValue("value2"),
							Recipe: nil,
						},
					},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			updatedStateSectionMap := toStateSectionsAndFieldsMap(tt.item, tt.stateSectionMap)

			if !reflect.DeepEqual(updatedStateSectionMap, tt.want) {
				t.Errorf("toStateSectionsAndFieldsMap() = %+v, want %+v", updatedStateSectionMap, tt.want)
			}
		})
	}
}
