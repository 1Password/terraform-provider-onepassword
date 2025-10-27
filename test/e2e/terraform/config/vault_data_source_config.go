package terraform

import "fmt"

type VaultDataSource struct {
	Params map[string]string
}

func VaultDataSourceConfig(params map[string]string) func() string {
	return func() string {
		dataSourceStr := `data "onepassword_vault" "test_vault" {`

		for key, value := range params {
			dataSourceStr += fmt.Sprintf("\n%s=\"%s\"", key, value)
		}

		dataSourceStr += "\n}"

		return dataSourceStr
	}
}
