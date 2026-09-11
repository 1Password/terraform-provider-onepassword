package terraform

import (
	"fmt"
	"strconv"
	"strings"
)

func VaultResourceConfig(name, description string, allowAdminsAccess bool, groupID string, permissions []string) func() string {
	return func() string {
		var config strings.Builder
		fmt.Fprintf(&config, `resource "onepassword_vault" "test_vault" {
  name                = %s
  description         = %s
  allow_admins_access = %t
`, strconv.Quote(name), strconv.Quote(description), allowAdminsAccess)

		if groupID != "" {
			config.WriteString("\n  group_permissions {\n")
			fmt.Fprintf(&config, "    group_id = %s\n", strconv.Quote(groupID))
			config.WriteString("    permissions = [")
			for i, permission := range permissions {
				if i > 0 {
					config.WriteString(", ")
				}
				config.WriteString(strconv.Quote(permission))
			}
			config.WriteString("]\n  }\n")
		}
		config.WriteString("}\n")
		return config.String()
	}
}
