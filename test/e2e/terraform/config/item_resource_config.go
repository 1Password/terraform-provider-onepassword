package terraform

import "fmt"

type ItemResource struct {
	Params map[string]string
}

func ItemResourceConfig(vaultID string, params map[string]string) func() string {
	return func() string {
		dataSourceStr := `resource "onepassword_item" "test_item" {`

		dataSourceStr += fmt.Sprintf("\n  vault = %q", vaultID)

		for key, value := range params {
			dataSourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		dataSourceStr += "\n}"
		return dataSourceStr
	}
}
