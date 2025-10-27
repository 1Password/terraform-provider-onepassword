// This package is used to provide helpers to generate terraform configuration files
package terraform

import (
	"fmt"
)

func ProviderAuthWithServiceAccount(config AuthConfig) func() string {
	return func() string {
		return fmt.Sprintf(`
		provider "onepassword" {
		  service_account_token = "%s"
		}
		`, config.ServiceAccountToken,
		)
	}
}

func ProviderAuthWithConnect(config ItemDataSource) string {
	return fmt.Sprintf(`
		provider "onepassword" {
			token = "%s"
			url = "%s"
		}
		`, config.Auth.ConnectToken, config.Auth.ConnectHost,
	)
}

type AuthConfig struct {
	ServiceAccountToken string
	ConnectHost         string
	ConnectToken        string
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

func CreateItemResourceConfigBuilder() func(functions ...func() string) string {
	configStr := ""

	return func(functions ...func() string) string {
		for _, f := range functions {
			configStr += f()
			configStr += "\n"
		}

		return configStr
	}
}
