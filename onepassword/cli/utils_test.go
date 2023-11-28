package cli

import (
	"reflect"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
)

func TestPasswordField(t *testing.T) {
	tests := map[string]struct {
		item          *onepassword.Item
		expectedField *onepassword.ItemField
	}{
		"should return nil if item has no fields": {
			item:          &onepassword.Item{},
			expectedField: nil,
		},
		"should return nil if no password field": {
			item: &onepassword.Item{
				Fields: []*onepassword.ItemField{
					{Purpose: onepassword.FieldPurposeNotes},
				},
			},
			expectedField: nil,
		},
		"should return password field": {
			item: &onepassword.Item{
				Fields: []*onepassword.ItemField{
					{ID: "username", Purpose: onepassword.FieldPurposeUsername},
					{ID: "password", Purpose: onepassword.FieldPurposePassword},
					{ID: "notes", Purpose: onepassword.FieldPurposeNotes},
				},
			},
			expectedField: &onepassword.ItemField{
				ID:      "password",
				Purpose: onepassword.FieldPurposePassword,
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			f := passwordField(test.item)

			if !reflect.DeepEqual(f, test.expectedField) {
				t.Errorf("Expected to \"%+v\" field, but got \"%+v\"", *test.expectedField, *f)
			}
		})
	}
}

func TestPasswordRecipeToString(t *testing.T) {
	tests := map[string]struct {
		recipe         *onepassword.GeneratorRecipe
		expectedString string
	}{
		"should return empty string if recipe is nil": {
			recipe:         nil,
			expectedString: "",
		},
		"should return empty string if recipe is default": {
			recipe:         &onepassword.GeneratorRecipe{},
			expectedString: "",
		},
		"should contain expected length": {
			recipe: &onepassword.GeneratorRecipe{
				Length: 30,
			},
			expectedString: "30",
		},
		"should contain letters charset": {
			recipe: &onepassword.GeneratorRecipe{
				CharacterSets: []string{"letters"},
			},
			expectedString: "letters",
		},
		"should contain letters and digits charsets": {
			recipe: &onepassword.GeneratorRecipe{
				CharacterSets: []string{"letters", "digits"},
			},
			expectedString: "letters,digits",
		},
		"should contain letters and digits charsets and length": {
			recipe: &onepassword.GeneratorRecipe{
				Length:        30,
				CharacterSets: []string{"letters", "digits"},
			},
			expectedString: "letters,digits,30",
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actualString := passwordRecipeToString(test.recipe)
			if actualString != test.expectedString {
				t.Errorf("Unexpected password recipe string. Expected \"%s\", but got \"%s\"", test.expectedString, actualString)
			}
		})
	}
}
