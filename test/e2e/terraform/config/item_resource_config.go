package terraform

import "fmt"

type ItemResource struct {
	Params map[string]string
}

type PasswordRecipe struct {
	Length  int
	Letters bool
	Digits  bool
	Symbols bool
}

func ItemResourceConfig(vaultID string, params map[string]string, passwordRecipe *PasswordRecipe) func() string {
	return func() string {
		resourceStr := `resource "onepassword_item" "test_item" {`

		resourceStr += fmt.Sprintf("\n  vault = %q", vaultID)

		for key, value := range params {
			if key == "tags" {
				resourceStr += fmt.Sprintf("\n  %s = [%q]", key, value)
				continue
			}

			if key == "password" && passwordRecipe != nil {
				continue
			}

			resourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		if passwordRecipe != nil {
			resourceStr += fmt.Sprintf(`
			password_recipe {
			length  = %d
			letters = %t
			digits  = %t
			symbols = %t
  			}`, passwordRecipe.Length, passwordRecipe.Letters, passwordRecipe.Digits, passwordRecipe.Symbols)
		}

		resourceStr += "\n}"
		return resourceStr
	}
}
