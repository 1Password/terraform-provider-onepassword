// test/e2e/terraform/config/ephemeral_item_config.go
package terraform

import "fmt"

type EphemeralItem struct {
	Params map[string]string
}

func EphemeralItemConfig(params map[string]string) func() string {
	return func() string {
		ephemeralStr := `ephemeral "onepassword_item" "test_item" {`

		for key, value := range params {
			ephemeralStr += fmt.Sprintf("\n%s=\"%s\"", key, value)
		}

		ephemeralStr += "\n}"

		return ephemeralStr
	}
}
