package terraform

import "fmt"

type ItemResource struct {
	Params map[string]string
	Tags   []string
}

func ItemResourceConfig(params map[string]string) func() string {
	return func() string {
		dataSourceStr := `resource "onepassword_item" "test_item" {`

		for key, value := range params {
			dataSourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		dataSourceStr += "\n}"
		return dataSourceStr
	}
}
