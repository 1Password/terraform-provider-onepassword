// This package is used to provide helpers to generate terraform configuration files
package terraform

func ProviderConfig() func() string {
	return func() string {
		return `
		provider "onepassword" {}
		`
	}
}

func CreateItemDataSourceConfigBuilder() func(functions ...func() string) string {
	configStr := ""

	return func(functions ...func() string) string {
		for _, f := range functions {
			configStr += f()
			configStr += "\n"
		}

		return configStr
	}
}
