package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

func TestToModelLoginFields(t *testing.T) {
	tests := map[string]struct {
		state    OnePasswordItemResourceModel
		password string
		recipe   *model.GeneratorRecipe
		want     []model.ItemField
	}{
		"with all fields": {
			state: OnePasswordItemResourceModel{
				Username:  types.StringValue("testuser"),
				NoteValue: types.StringValue("test notes"),
			},
			password: "testpass",
			recipe:   nil,
			want: []model.ItemField{
				{
					ID:      "username",
					Label:   "username",
					Purpose: model.FieldPurposeUsername,
					Type:    model.FieldTypeString,
					Value:   "testuser",
				},
				{
					ID:       "password",
					Label:    "password",
					Purpose:  model.FieldPurposePassword,
					Type:     model.FieldTypeConcealed,
					Value:    "testpass",
					Generate: false,
					Recipe:   nil,
				},
				{
					ID:      "notesPlain",
					Label:   "notesPlain",
					Type:    model.FieldTypeString,
					Purpose: model.FieldPurposeNotes,
					Value:   "test notes",
				},
			},
		},
		"with empty password should generate": {
			state: OnePasswordItemResourceModel{
				Username:  types.StringValue("testuser"),
				NoteValue: types.StringValue("test notes"),
			},
			password: "",
			recipe: &model.GeneratorRecipe{
				Length:        32,
				CharacterSets: []model.CharacterSet{model.CharacterSetDigits},
			},
			want: []model.ItemField{
				{
					ID:      "username",
					Label:   "username",
					Purpose: model.FieldPurposeUsername,
					Type:    model.FieldTypeString,
					Value:   "testuser",
				},
				{
					ID:       "password",
					Label:    "password",
					Purpose:  model.FieldPurposePassword,
					Type:     model.FieldTypeConcealed,
					Value:    "",
					Generate: true,
					Recipe: &model.GeneratorRecipe{
						Length:        32,
						CharacterSets: []model.CharacterSet{model.CharacterSetDigits},
					},
				},
				{
					ID:      "notesPlain",
					Label:   "notesPlain",
					Type:    model.FieldTypeString,
					Purpose: model.FieldPurposeNotes,
					Value:   "test notes",
				},
			},
		},
		"with null values": {
			state: OnePasswordItemResourceModel{
				Username:  types.StringNull(),
				NoteValue: types.StringNull(),
			},
			password: "testpass",
			recipe:   nil,
			want: []model.ItemField{
				{
					ID:      "username",
					Label:   "username",
					Purpose: model.FieldPurposeUsername,
					Type:    model.FieldTypeString,
					Value:   "",
				},
				{
					ID:       "password",
					Label:    "password",
					Purpose:  model.FieldPurposePassword,
					Type:     model.FieldTypeConcealed,
					Value:    "testpass",
					Generate: false,
					Recipe:   nil,
				},
				{
					ID:      "notesPlain",
					Label:   "notesPlain",
					Type:    model.FieldTypeString,
					Purpose: model.FieldPurposeNotes,
					Value:   "",
				},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			got := toModelLoginFields(test.state, test.password, test.recipe)
			if len(got) != len(test.want) {
				t.Errorf("Field count mismatch: got %d, want %d", len(got), len(test.want))
				return
			}
			for i, wantField := range test.want {
				if got[i].ID != wantField.ID {
					t.Errorf("Field[%d].ID: got %v, want %v", i, got[i].ID, wantField.ID)
				}
				if got[i].Label != wantField.Label {
					t.Errorf("Field[%d].Label: got %v, want %v", i, got[i].Label, wantField.Label)
				}
				if got[i].Type != wantField.Type {
					t.Errorf("Field[%d].Type: got %v, want %v", i, got[i].Type, wantField.Type)
				}
				if got[i].Purpose != wantField.Purpose {
					t.Errorf("Field[%d].Purpose: got %v, want %v", i, got[i].Purpose, wantField.Purpose)
				}
				if got[i].Value != wantField.Value {
					t.Errorf("Field[%d].Value: got %v, want %v", i, got[i].Value, wantField.Value)
				}
				if got[i].Generate != wantField.Generate {
					t.Errorf("Field[%d].Generate: got %v, want %v", i, got[i].Generate, wantField.Generate)
				}
				if (got[i].Recipe == nil) != (wantField.Recipe == nil) {
					t.Errorf("Field[%d].Recipe: got nil=%v, want nil=%v", i, got[i].Recipe == nil, wantField.Recipe == nil)
				} else if got[i].Recipe != nil && wantField.Recipe != nil {
					if got[i].Recipe.Length != wantField.Recipe.Length {
						t.Errorf("Field[%d].Recipe.Length: got %v, want %v", i, got[i].Recipe.Length, wantField.Recipe.Length)
					}
					if !reflect.DeepEqual(got[i].Recipe.CharacterSets, wantField.Recipe.CharacterSets) {
						t.Errorf("Field[%d].Recipe.CharacterSets: got %v, want %v", i, got[i].Recipe.CharacterSets, wantField.Recipe.CharacterSets)
					}
				}
			}
		})
	}
}

func TestToModelPasswordFields(t *testing.T) {
	tests := map[string]struct {
		state    OnePasswordItemResourceModel
		password string
		recipe   *model.GeneratorRecipe
		wantLen  int
		wantIDs  []string
	}{
		"with password and notes": {
			state: OnePasswordItemResourceModel{
				NoteValue: types.StringValue("test notes"),
			},
			password: "testpass",
			recipe:   nil,
			wantLen:  2,
			wantIDs:  []string{"password", "notesPlain"},
		},
		"with empty password": {
			state: OnePasswordItemResourceModel{
				NoteValue: types.StringValue("test notes"),
			},
			password: "",
			recipe: &model.GeneratorRecipe{
				Length: 32,
			},
			wantLen: 2,
			wantIDs: []string{"password", "notesPlain"},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			got := toModelPasswordFields(test.state, test.password, test.recipe)
			if len(got) != test.wantLen {
				t.Errorf("Field count: got %d, want %d", len(got), test.wantLen)
				return
			}
			for i, wantID := range test.wantIDs {
				if got[i].ID != wantID {
					t.Errorf("Field[%d].ID: got %v, want %v", i, got[i].ID, wantID)
				}
			}
			if got[0].Purpose != model.FieldPurposePassword {
				t.Errorf("Field[0].Purpose: got %v, want %v", got[0].Purpose, model.FieldPurposePassword)
			}
			if got[1].Purpose != model.FieldPurposeNotes {
				t.Errorf("Field[1].Purpose: got %v, want %v", got[1].Purpose, model.FieldPurposeNotes)
			}
		})
	}
}

func TestToModelDatabaseFields(t *testing.T) {
	state := OnePasswordItemResourceModel{
		Username:  types.StringValue("dbuser"),
		Hostname:  types.StringValue("dbhost"),
		Database:  types.StringValue("mydb"),
		Port:      types.StringValue("3306"),
		Type:      types.StringValue("mysql"),
		NoteValue: types.StringValue("database notes"),
	}
	password := "dbpass"
	recipe := &model.GeneratorRecipe{Length: 32}

	got := toModelDatabaseFields(state, password, recipe)

	if len(got) != 7 {
		t.Fatalf("Field count: got %d, want 7", len(got))
	}

	// Check username field
	if got[0].ID != "username" || got[0].Value != "dbuser" {
		t.Errorf("Username field: got ID=%v Value=%v, want ID=username Value=dbuser", got[0].ID, got[0].Value)
	}

	// Check password field
	if got[1].ID != "password" || got[1].Value != "dbpass" || got[1].Generate {
		t.Errorf("Password field: got ID=%v Value=%v Generate=%v, want ID=password Value=dbpass Generate=false", got[1].ID, got[1].Value, got[1].Generate)
	}

	// Check hostname field
	if got[2].ID != "hostname" || got[2].Value != "dbhost" {
		t.Errorf("Hostname field: got ID=%v Value=%v, want ID=hostname Value=dbhost", got[2].ID, got[2].Value)
	}

	// Check database field
	if got[3].ID != "database" || got[3].Value != "mydb" {
		t.Errorf("Database field: got ID=%v Value=%v, want ID=database Value=mydb", got[3].ID, got[3].Value)
	}

	// Check port field
	if got[4].ID != "port" || got[4].Value != "3306" {
		t.Errorf("Port field: got ID=%v Value=%v, want ID=port Value=3306", got[4].ID, got[4].Value)
	}

	// Check type field
	if got[5].ID != "database_type" || got[5].Label != "type" || got[5].Value != "mysql" {
		t.Errorf("Type field: got ID=%v Label=%v Value=%v, want ID=database_type Label=type Value=mysql", got[5].ID, got[5].Label, got[5].Value)
	}

	// Check notes field
	if got[6].ID != "notesPlain" || got[6].Purpose != model.FieldPurposeNotes || got[6].Value != "database notes" {
		t.Errorf("Notes field: got ID=%v Purpose=%v Value=%v, want ID=notesPlain Purpose=%v Value=database notes", got[6].ID, got[6].Purpose, got[6].Value, model.FieldPurposeNotes)
	}
}

func TestToModelSecureNoteFields(t *testing.T) {
	tests := map[string]struct {
		state    OnePasswordItemResourceModel
		wantNote string
	}{
		"with notes": {
			state: OnePasswordItemResourceModel{
				NoteValue: types.StringValue("secure note content"),
			},
			wantNote: "secure note content",
		},
		"with empty notes": {
			state: OnePasswordItemResourceModel{
				NoteValue: types.StringValue(""),
			},
			wantNote: "",
		},
		"with null notes": {
			state: OnePasswordItemResourceModel{
				NoteValue: types.StringNull(),
			},
			wantNote: "",
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			got := toModelSecureNoteFields(test.state)
			if len(got) != 1 {
				t.Fatalf("Field count: got %d, want 1", len(got))
			}
			if got[0].ID != "notesPlain" {
				t.Errorf("Field.ID: got %v, want notesPlain", got[0].ID)
			}
			if got[0].Purpose != model.FieldPurposeNotes {
				t.Errorf("Field.Purpose: got %v, want %v", got[0].Purpose, model.FieldPurposeNotes)
			}
			if got[0].Type != model.FieldTypeString {
				t.Errorf("Field.Type: got %v, want %v", got[0].Type, model.FieldTypeString)
			}
			if got[0].Value != test.wantNote {
				t.Errorf("Field.Value: got %v, want %v", got[0].Value, test.wantNote)
			}
		})
	}
}

func TestToModelSectionField(t *testing.T) {
	tests := map[string]struct {
		state        OnePasswordItemResourceFieldModel
		sectionID    string
		sectionLabel string
		wantErr      bool
		validate     func(t *testing.T, field *model.ItemField)
	}{
		"with existing field ID": {
			state: OnePasswordItemResourceFieldModel{
				ID:      types.StringValue("existing-field-id"),
				Label:   types.StringValue("Test Field"),
				Type:    types.StringValue("STRING"),
				Purpose: types.StringValue("USERNAME"),
				Value:   types.StringValue("test value"),
				Recipe:  []PasswordRecipeModel{},
			},
			sectionID:    "section-id",
			sectionLabel: "Section Label",
			wantErr:      false,
			validate: func(t *testing.T, field *model.ItemField) {
				if field.ID != "existing-field-id" {
					t.Errorf("Field.ID: got %v, want existing-field-id", field.ID)
				}
				if field.SectionID != "section-id" {
					t.Errorf("Field.SectionID: got %v, want section-id", field.SectionID)
				}
				if field.SectionLabel != "Section Label" {
					t.Errorf("Field.SectionLabel: got %v, want Section Label", field.SectionLabel)
				}
				if field.Label != "Test Field" {
					t.Errorf("Field.Label: got %v, want Test Field", field.Label)
				}
				if field.Value != "test value" {
					t.Errorf("Field.Value: got %v, want test value", field.Value)
				}
				if field.Type != model.ItemFieldType("STRING") {
					t.Errorf("Field.Type: got %v, want STRING", field.Type)
				}
				if field.Purpose != model.ItemFieldPurpose("USERNAME") {
					t.Errorf("Field.Purpose: got %v, want USERNAME", field.Purpose)
				}
			},
		},
		"without field ID generates UUID": {
			state: OnePasswordItemResourceFieldModel{
				ID:      types.StringValue(""),
				Label:   types.StringValue("Test Field"),
				Type:    types.StringValue("CONCEALED"),
				Purpose: types.StringValue("PASSWORD"),
				Value:   types.StringValue("secret"),
				Recipe:  []PasswordRecipeModel{},
			},
			sectionID:    "section-id",
			sectionLabel: "Section Label",
			wantErr:      false,
			validate: func(t *testing.T, field *model.ItemField) {
				if field.ID == "" {
					t.Error("Field.ID: should generate UUID, got empty string")
				}
				if field.SectionID != "section-id" {
					t.Errorf("Field.SectionID: got %v, want section-id", field.SectionID)
				}
				if field.Type != model.ItemFieldType("CONCEALED") {
					t.Errorf("Field.Type: got %v, want CONCEALED", field.Type)
				}
			},
		},
		"with recipe": {
			state: OnePasswordItemResourceFieldModel{
				ID:      types.StringValue("field-id"),
				Label:   types.StringValue("Password Field"),
				Type:    types.StringValue("CONCEALED"),
				Purpose: types.StringValue("PASSWORD"),
				Value:   types.StringValue(""),
				Recipe: []PasswordRecipeModel{
					{
						Length:  types.Int64Value(16),
						Digits:  types.BoolValue(true),
						Symbols: types.BoolValue(true),
					},
				},
			},
			sectionID:    "section-id",
			sectionLabel: "Section Label",
			wantErr:      false,
			validate: func(t *testing.T, field *model.ItemField) {
				if field.Recipe == nil {
					t.Error("Field.Recipe: should not be nil")
				} else if field.Recipe.Length != 16 {
					t.Errorf("Field.Recipe.Length: got %v, want 16", field.Recipe.Length)
				}
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			got, err := toModelSectionField(test.state, test.sectionID, test.sectionLabel)
			if (err != nil) != test.wantErr {
				t.Errorf("Error: got err=%v, wantErr=%v", err != nil, test.wantErr)
				return
			}
			if test.wantErr {
				if got != nil {
					t.Errorf("Field: got %v, want nil on error", got)
				}
				return
			}
			if got == nil {
				t.Fatal("Field: got nil, want non-nil")
			}
			if test.validate != nil {
				test.validate(t, got)
			}
		})
	}
}

func TestToModelSections(t *testing.T) {
	tests := map[string]struct {
		state    OnePasswordItemResourceModel
		wantErr  bool
		validate func(t *testing.T, item *model.Item)
	}{
		"with sections and fields": {
			state: OnePasswordItemResourceModel{
				Section: []OnePasswordItemResourceSectionModel{
					{
						ID:    types.StringValue("section-1"),
						Label: types.StringValue("Test Section"),
						Field: []OnePasswordItemResourceFieldModel{
							{
								ID:      types.StringValue("field-1"),
								Label:   types.StringValue("Field 1"),
								Type:    types.StringValue("STRING"),
								Purpose: types.StringValue(""),
								Value:   types.StringValue("value 1"),
								Recipe:  []PasswordRecipeModel{},
							},
							{
								ID:      types.StringValue("field-2"),
								Label:   types.StringValue("Field 2"),
								Type:    types.StringValue("CONCEALED"),
								Purpose: types.StringValue("PASSWORD"),
								Value:   types.StringValue("secret"),
								Recipe:  []PasswordRecipeModel{},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, item *model.Item) {
				if len(item.Sections) != 1 {
					t.Fatalf("Sections count: got %d, want 1", len(item.Sections))
				}
				if item.Sections[0].ID != "section-1" {
					t.Errorf("Section[0].ID: got %v, want section-1", item.Sections[0].ID)
				}
				if item.Sections[0].Label != "Test Section" {
					t.Errorf("Section[0].Label: got %v, want Test Section", item.Sections[0].Label)
				}
				if len(item.Fields) != 2 {
					t.Fatalf("Fields count: got %d, want 2", len(item.Fields))
				}
				if item.Fields[0].ID != "field-1" {
					t.Errorf("Field[0].ID: got %v, want field-1", item.Fields[0].ID)
				}
				if item.Fields[1].ID != "field-2" {
					t.Errorf("Field[1].ID: got %v, want field-2", item.Fields[1].ID)
				}
				if item.Fields[0].SectionID != "section-1" {
					t.Errorf("Field[0].SectionID: got %v, want section-1", item.Fields[0].SectionID)
				}
				if item.Fields[1].SectionID != "section-1" {
					t.Errorf("Field[1].SectionID: got %v, want section-1", item.Fields[1].SectionID)
				}
			},
		},
		"with section without ID generates UUID": {
			state: OnePasswordItemResourceModel{
				Section: []OnePasswordItemResourceSectionModel{
					{
						ID:    types.StringValue(""),
						Label: types.StringValue("New Section"),
						Field: []OnePasswordItemResourceFieldModel{
							{
								ID:     types.StringValue("field-1"),
								Label:  types.StringValue("Field 1"),
								Type:   types.StringValue("STRING"),
								Value:  types.StringValue("value"),
								Recipe: []PasswordRecipeModel{},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, item *model.Item) {
				if len(item.Sections) != 1 {
					t.Fatalf("Sections count: got %d, want 1", len(item.Sections))
				}
				if item.Sections[0].ID == "" {
					t.Error("Section[0].ID: should generate UUID, got empty string")
				}
				if item.Sections[0].Label != "New Section" {
					t.Errorf("Section[0].Label: got %v, want New Section", item.Sections[0].Label)
				}
				if len(item.Fields) != 1 {
					t.Fatalf("Fields count: got %d, want 1", len(item.Fields))
				}
				if item.Fields[0].SectionID != item.Sections[0].ID {
					t.Errorf("Field[0].SectionID: got %v, want %v", item.Fields[0].SectionID, item.Sections[0].ID)
				}
			},
		},
		"with multiple sections": {
			state: OnePasswordItemResourceModel{
				Section: []OnePasswordItemResourceSectionModel{
					{
						ID:    types.StringValue("section-1"),
						Label: types.StringValue("Section 1"),
						Field: []OnePasswordItemResourceFieldModel{
							{
								ID:     types.StringValue("field-1"),
								Label:  types.StringValue("Field 1"),
								Type:   types.StringValue("STRING"),
								Value:  types.StringValue("value 1"),
								Recipe: []PasswordRecipeModel{},
							},
						},
					},
					{
						ID:    types.StringValue("section-2"),
						Label: types.StringValue("Section 2"),
						Field: []OnePasswordItemResourceFieldModel{
							{
								ID:     types.StringValue("field-2"),
								Label:  types.StringValue("Field 2"),
								Type:   types.StringValue("STRING"),
								Value:  types.StringValue("value 2"),
								Recipe: []PasswordRecipeModel{},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, item *model.Item) {
				if len(item.Sections) != 2 {
					t.Fatalf("Sections count: got %d, want 2", len(item.Sections))
				}
				if len(item.Fields) != 2 {
					t.Fatalf("Fields count: got %d, want 2", len(item.Fields))
				}
				if item.Fields[0].SectionID != "section-1" {
					t.Errorf("Field[0].SectionID: got %v, want section-1", item.Fields[0].SectionID)
				}
				if item.Fields[1].SectionID != "section-2" {
					t.Errorf("Field[1].SectionID: got %v, want section-2", item.Fields[1].SectionID)
				}
			},
		},
		"with empty sections": {
			state: OnePasswordItemResourceModel{
				Section: []OnePasswordItemResourceSectionModel{},
			},
			wantErr: false,
			validate: func(t *testing.T, item *model.Item) {
				if len(item.Sections) != 0 {
					t.Errorf("Sections count: got %d, want 0", len(item.Sections))
				}
				if len(item.Fields) != 0 {
					t.Errorf("Fields count: got %d, want 0", len(item.Fields))
				}
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			item := &model.Item{}
			err := toModelSections(test.state, item)
			if (err != nil) != test.wantErr {
				t.Errorf("Error: got err=%v, wantErr=%v", err != nil, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if test.validate != nil {
				test.validate(t, item)
			}
		})
	}
}

func TestToModelTags(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		state    OnePasswordItemResourceModel
		wantTags []string
		wantErr  bool
	}{
		"with tags": {
			state: func() OnePasswordItemResourceModel {
				tags, _ := types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2", "tag3"})
				return OnePasswordItemResourceModel{Tags: tags}
			}(),
			wantTags: []string{"tag1", "tag2", "tag3"},
			wantErr:  false,
		},
		"with empty tags": {
			state: func() OnePasswordItemResourceModel {
				tags, _ := types.ListValueFrom(ctx, types.StringType, []string{})
				return OnePasswordItemResourceModel{Tags: tags}
			}(),
			wantTags: []string{},
			wantErr:  false,
		},
		"with null tags": {
			state: OnePasswordItemResourceModel{
				Tags: types.ListNull(types.StringType),
			},
			wantTags: []string{},
			wantErr:  false,
		},
		"with single tag": {
			state: func() OnePasswordItemResourceModel {
				tags, _ := types.ListValueFrom(ctx, types.StringType, []string{"single"})
				return OnePasswordItemResourceModel{Tags: tags}
			}(),
			wantTags: []string{"single"},
			wantErr:  false,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			got, err := toModelTags(ctx, test.state)
			if (err != nil && err.HasError()) != test.wantErr {
				t.Errorf("Error: got err=%v, wantErr=%v", err != nil && err.HasError(), test.wantErr)
				return
			}
			if test.wantErr {
				if got != nil {
					t.Errorf("Tags: got %v, want nil on error", got)
				}
				return
			}

			if got == nil {
				got = []string{}
			}
			if test.wantTags == nil {
				test.wantTags = []string{}
			}
			if !reflect.DeepEqual(got, test.wantTags) {
				t.Errorf("Tags: got %v, want %v", got, test.wantTags)
			}
		})
	}
}
