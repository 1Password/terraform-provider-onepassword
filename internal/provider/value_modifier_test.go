package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueFitsRecipe(t *testing.T) {
	recipe := func(length int64, digits, symbols bool) PasswordRecipeModel {
		return PasswordRecipeModel{
			Length:  types.Int64Value(length),
			Digits:  types.BoolValue(digits),
			Symbols: types.BoolValue(symbols),
		}
	}

	testCases := map[string]struct {
		value  string
		recipe PasswordRecipeModel
		want   bool
	}{
		"fits":                       {"aB3!cD4@eF5.", recipe(12, true, true), true},
		"wrong length":               {"aB3!cD4@", recipe(12, true, true), false},
		"length counts characters":   {"aB3!cD4é", recipe(8, true, true), true},
		"digits when digits=false":   {"abcd1234", recipe(8, false, false), false},
		"symbols when symbols=false": {"abcd!@#$", recipe(8, false, false), false},
		"no digits when digits=true": {"abcd!@#$", recipe(8, true, true), false},
		"unknown length": {"abcdEFGH", PasswordRecipeModel{
			Length: types.Int64Unknown(), Digits: types.BoolValue(false), Symbols: types.BoolValue(false),
		}, false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := valueFitsRecipe(tc.value, tc.recipe); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
