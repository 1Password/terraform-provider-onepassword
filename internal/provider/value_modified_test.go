package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestShouldPreservePasswordValue(t *testing.T) {
	recipe32 := []PasswordRecipeModel{{
		Length:  types.Int64Value(32),
		Digits:  types.BoolValue(true),
		Symbols: types.BoolValue(true),
	}}
	recipe20 := []PasswordRecipeModel{{
		Length:  types.Int64Value(20),
		Digits:  types.BoolValue(true),
		Symbols: types.BoolValue(false),
	}}

	tests := []struct {
		name                      string
		stateRecipe               []PasswordRecipeModel
		planRecipe                []PasswordRecipeModel
		passwordShouldBePreserved bool
	}{
		{
			name:                      "when state and plan recipe are both nil then preserve password",
			stateRecipe:               nil,
			planRecipe:                nil,
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state and plan recipe are both empty slice then preserve password",
			stateRecipe:               []PasswordRecipeModel{},
			planRecipe:                []PasswordRecipeModel{},
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state and plan have same recipe then preserve password",
			stateRecipe:               recipe32,
			planRecipe:                recipe32,
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state has no recipe and plan has recipe (post-import) then preserve password",
			stateRecipe:               nil,
			planRecipe:                recipe32,
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state recipe is empty and plan has recipe then preserve password",
			stateRecipe:               []PasswordRecipeModel{},
			planRecipe:                recipe32,
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state and plan have different recipes then do not preserve password",
			stateRecipe:               recipe32,
			planRecipe:                recipe20,
			passwordShouldBePreserved: false,
		},
		{
			name:                      "when state has recipe and plan has no recipe then do not preserve password",
			stateRecipe:               recipe32,
			planRecipe:                nil,
			passwordShouldBePreserved: false,
		},
		{
			name:                      "when state has recipe and plan recipe is empty then do not preserve password",
			stateRecipe:               recipe32,
			planRecipe:                []PasswordRecipeModel{},
			passwordShouldBePreserved: false,
		},
		{
			name:                      "when recipes have same length but different digits or symbols then do not preserve password",
			stateRecipe:               []PasswordRecipeModel{{Length: types.Int64Value(32), Digits: types.BoolValue(true), Symbols: types.BoolValue(true)}},
			planRecipe:                []PasswordRecipeModel{{Length: types.Int64Value(32), Digits: types.BoolValue(false), Symbols: types.BoolValue(true)}},
			passwordShouldBePreserved: false,
		},
		{
			name:                      "when state has no recipe and plan has recipe with zero values then preserve password",
			stateRecipe:               nil,
			planRecipe:                []PasswordRecipeModel{{Length: types.Int64Value(0), Digits: types.BoolValue(false), Symbols: types.BoolValue(false)}},
			passwordShouldBePreserved: true,
		},
		{
			name:                      "when state is nil and plan is empty slice then do not preserve password",
			stateRecipe:               nil,
			planRecipe:                []PasswordRecipeModel{},
			passwordShouldBePreserved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPreservePasswordValue(tt.stateRecipe, tt.planRecipe)
			if got != tt.passwordShouldBePreserved {
				t.Errorf("shouldPreservePasswordValue() = %v, want password preserved = %v", got, tt.passwordShouldBePreserved)
			}
		})
	}

}
