package terraform

import "fmt"

type ItemResource struct {
	Auth   AuthConfig
	Params map[string]string
	Tags   []string
}

func ItemResourceConfig(params map[string]string, tags []string) func() string {
	return func() string {
		dataSourceStr := `resource "onepassword_item" "test_item" {`

		for key, value := range params {
			dataSourceStr += fmt.Sprintf("\n  %s = %q", key, value)
		}

		// Add tags if they exist
		if len(tags) > 0 {
			dataSourceStr += "\n  tags = ["
			for i, tag := range tags {
				if i > 0 {
					dataSourceStr += ", "
				}
				dataSourceStr += fmt.Sprintf("%q", tag)
			}
			dataSourceStr += "]"
		}

		dataSourceStr += "\n}"
		return dataSourceStr
	}
}
