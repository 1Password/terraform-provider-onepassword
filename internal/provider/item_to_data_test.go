package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProcessTags(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		itemTags    []string
		currentTags types.List
		want        types.List
		wantErr     bool
	}{
		{
			name:     "tags match",
			itemTags: []string{"tag1", "tag2"},
		},
		{
			name:     "tags differ",
			itemTags: []string{"tag1", "tag2"},
		},
		{
			name:        "empty tags preserve null",
			itemTags:    []string{},
			currentTags: types.ListNull(types.StringType),
			want:        types.ListNull(types.StringType),
		},
		{
			name:     "empty tags with existing list",
			itemTags: []string{},
		},
		{
			name:        "null current tags",
			itemTags:    []string{"tag1"},
			currentTags: types.ListNull(types.StringType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test-specific values
			var currentTags, want types.List
			switch tt.name {
			case "tags match":
				currentTags, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})
				want, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})
			case "tags differ":
				currentTags, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag3"})
				want, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})
			case "empty tags preserve null":
				currentTags = tt.currentTags
				want = tt.want
			case "empty tags with existing list":
				currentTags, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag1"})
				want, _ = types.ListValueFrom(ctx, types.StringType, []string{})
			case "null current tags":
				currentTags = tt.currentTags
				want, _ = types.ListValueFrom(ctx, types.StringType, []string{"tag1"})
			}

			got, diags := processTags(ctx, tt.itemTags, currentTags)
			if (diags.HasError()) != tt.wantErr {
				t.Errorf("processTags() error = %v, wantErr %v", diags.HasError(), tt.wantErr)
				return
			}
			if !tt.wantErr {
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

func TestProcessSectionsAndFields(t *testing.T) {
	tests := []struct {
		name         string
		itemSections []model.ItemSection
		itemFields   []model.ItemField
		dataSections []OnePasswordItemResourceSectionModel
		want         []OnePasswordItemResourceSectionModel
	}{
		{
			name: "new section with field",
			itemSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			itemFields: []model.ItemField{
				{ID: "field1", Label: "Field 1", Type: model.FieldTypeString, Value: "value1", SectionID: "section1"},
			},
			dataSections: []OnePasswordItemResourceSectionModel{},
			want: []OnePasswordItemResourceSectionModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					Field: []OnePasswordItemResourceFieldModel{
						{
							ID:      types.StringValue("field1"),
							Label:   types.StringValue("Field 1"),
							Type:    types.StringValue("STRING"),
							Value:   types.StringValue("value1"),
							Purpose: types.StringNull(),
						},
					},
				},
			},
		},
		{
			name: "update existing section by ID",
			itemSections: []model.ItemSection{
				{ID: "section1", Label: "Updated Label"},
			},
			itemFields: []model.ItemField{},
			dataSections: []OnePasswordItemResourceSectionModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Old Label"),
					Field: []OnePasswordItemResourceFieldModel{},
				},
			},
			want: []OnePasswordItemResourceSectionModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Updated Label"),
					Field: []OnePasswordItemResourceFieldModel{},
				},
			},
		},
		{
			name: "field with recipe",
			itemSections: []model.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			itemFields: []model.ItemField{
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
			dataSections: []OnePasswordItemResourceSectionModel{},
			want: []OnePasswordItemResourceSectionModel{
				{
					ID:    types.StringValue("section1"),
					Label: types.StringValue("Section 1"),
					Field: []OnePasswordItemResourceFieldModel{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processSectionsAndFields(tt.itemSections, tt.itemFields, tt.dataSections)
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
				if len(got[i].Field) != len(tt.want[i].Field) {
					t.Errorf("Section[%d].Field len = %d, want %d", i, len(got[i].Field), len(tt.want[i].Field))
				}
			}
		})
	}
}

func TestProcessTopLevelFields(t *testing.T) {
	tests := []struct {
		name       string
		itemFields []model.ItemField
		data       *OnePasswordItemResourceModel
		want       *OnePasswordItemResourceModel
	}{
		{
			name: "username by purpose",
			itemFields: []model.ItemField{
				{Purpose: model.FieldPurposeUsername, Value: "user1"},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Username: types.StringValue("user1"),
			},
		},
		{
			name: "password by purpose",
			itemFields: []model.ItemField{
				{Purpose: model.FieldPurposePassword, Value: "pass1"},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Password: types.StringValue("pass1"),
			},
		},
		{
			name: "notes by purpose",
			itemFields: []model.ItemField{
				{Purpose: model.FieldPurposeNotes, Value: "note1"},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				NoteValue: types.StringValue("note1"),
			},
		},
		{
			name: "hostname by label",
			itemFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: ""},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Hostname: types.StringValue("example.com"),
			},
		},
		{
			name: "database by label",
			itemFields: []model.ItemField{
				{Label: "database", Value: "mydb", SectionID: ""},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Database: types.StringValue("mydb"),
			},
		},
		{
			name: "port by label",
			itemFields: []model.ItemField{
				{Label: "port", Value: "3306", SectionID: ""},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Port: types.StringValue("3306"),
			},
		},
		{
			name: "type by label",
			itemFields: []model.ItemField{
				{Label: "type", Value: "mysql", SectionID: ""},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{
				Type: types.StringValue("mysql"),
			},
		},
		{
			name: "field in section ignored",
			itemFields: []model.ItemField{
				{Label: "hostname", Value: "example.com", SectionID: "section1"},
			},
			data: &OnePasswordItemResourceModel{},
			want: &OnePasswordItemResourceModel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processTopLevelFields(tt.itemFields, tt.data)
			if tt.data.Username.ValueString() != tt.want.Username.ValueString() {
				t.Errorf("Username = %v, want %v", tt.data.Username.ValueString(), tt.want.Username.ValueString())
			}
			if tt.data.Password.ValueString() != tt.want.Password.ValueString() {
				t.Errorf("Password = %v, want %v", tt.data.Password.ValueString(), tt.want.Password.ValueString())
			}
			if tt.data.NoteValue.ValueString() != tt.want.NoteValue.ValueString() {
				t.Errorf("NoteValue = %v, want %v", tt.data.NoteValue.ValueString(), tt.want.NoteValue.ValueString())
			}
			if tt.data.Hostname.ValueString() != tt.want.Hostname.ValueString() {
				t.Errorf("Hostname = %v, want %v", tt.data.Hostname.ValueString(), tt.want.Hostname.ValueString())
			}
			if tt.data.Database.ValueString() != tt.want.Database.ValueString() {
				t.Errorf("Database = %v, want %v", tt.data.Database.ValueString(), tt.want.Database.ValueString())
			}
			if tt.data.Port.ValueString() != tt.want.Port.ValueString() {
				t.Errorf("Port = %v, want %v", tt.data.Port.ValueString(), tt.want.Port.ValueString())
			}
			if tt.data.Type.ValueString() != tt.want.Type.ValueString() {
				t.Errorf("Type = %v, want %v", tt.data.Type.ValueString(), tt.want.Type.ValueString())
			}
		})
	}
}
