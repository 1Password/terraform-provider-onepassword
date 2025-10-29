package terraform

import "fmt"

type ItemResource struct {
	Params map[string]string
}

func ItemResourceConfig(vaultID string, params map[string]string, passwordRecipe bool) func() string {
	return func() string {
		dataSourceStr := `resource "onepassword_item" "test_item" {`

		dataSourceStr += fmt.Sprintf("\n  vault = %q", vaultID)

		for key, value := range params {
			if key == "tags" {
				dataSourceStr += fmt.Sprintf("\n  %s = [%q]", key, value)
				continue
			}
			dataSourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		if passwordRecipe {
			dataSourceStr += `
			password_recipe {
			length  = 40
			letters = true
			digits  = true
			symbols = false
  			}`
		}

		dataSourceStr += "\n}"
		return dataSourceStr
	}
}
