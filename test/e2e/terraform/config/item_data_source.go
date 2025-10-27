package terraform

import "fmt"

type ItemDataSource struct {
	Params map[string]string
}

func ItemDataSourceConfig(params map[string]string) func() string {
	return func() string {
		dataSourceStr := `data "onepassword_item" "test_item" {`

		for key, value := range params {
			dataSourceStr += fmt.Sprintf("\n%s=\"%s\"", key, value)
		}

		dataSourceStr += "\n}"

		return dataSourceStr
	}
}
