package terraform

import (
	"fmt"
	"reflect"
	"strings"
)

func ItemResourceConfig(vaultID string, params map[string]any) func() string {
	return func() string {
		resourceStr := `resource "onepassword_item" "test_item" {`

		resourceStr += fmt.Sprintf("\n  vault = %q", vaultID)

		for key, value := range params {
			resourceStr += formatTerraformAttribute(key, value)
		}

		resourceStr += "\n}"
		return resourceStr
	}
}

func formatTerraformAttribute(key string, value any) string {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		quotedItems := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			quotedItems[i] = fmt.Sprintf("%q", rv.Index(i).Interface())
		}
		return fmt.Sprintf("\n  %s = [%s]", key, strings.Join(quotedItems, ", "))

	case reflect.Bool:
		return fmt.Sprintf("\n  %s = %t", key, value)

	case reflect.String:
		return fmt.Sprintf("\n  %s = %q", key, value)

	case reflect.Int:
		return fmt.Sprintf("\n  %s = %d", key, value)

	default:
		return fmt.Sprintf("\n  %s = %q", key, value)
	}
}
