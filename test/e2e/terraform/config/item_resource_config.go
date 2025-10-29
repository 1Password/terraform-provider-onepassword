package terraform

import "fmt"

type ItemResource struct {
	Params map[string]string
}

func ItemResourceConfig(vaultID string, params map[string]string, passwordRecipe bool) func() string {
	return func() string {
		resourceStr := `resource "onepassword_item" "test_item" {`

		resourceStr += fmt.Sprintf("\n  vault = %q", vaultID)

		for key, value := range params {
			if key == "tags" {
				resourceStr += fmt.Sprintf("\n  %s = [%q]", key, value)
				continue
			}
			resourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		if passwordRecipe {
			resourceStr += `
			password_recipe {
			length  = 40
			letters = true
			digits  = true
			symbols = false
  			}`
		}

		resourceStr += "\n}"
		return resourceStr
	}
}
