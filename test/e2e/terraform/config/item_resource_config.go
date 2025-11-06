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
			attr, err := formatTerraformAttribute(key, value)
			if err != nil {
				return fmt.Sprintf("ERROR: %v", err)
			}
			resourceStr += attr
		}

		resourceStr += "\n}"
		return resourceStr
	}
}

func formatTerraformAttribute(key string, value any) (string, error) {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		// Handle slices of maps recursively
		if rv.Len() > 0 && rv.Index(0).Kind() == reflect.Map {
			blockStr := ""

			for i := 0; i < rv.Len(); i++ {
				blockStr += fmt.Sprintf("\n  %s {", key)
				attributes, ok := rv.Index(i).Interface().(map[string]any)

				if !ok {
					return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
				}

				for k, v := range attributes {
					attr, err := formatTerraformAttribute(k, v)
					if err != nil {
						return "", err
					}
					blockStr += attr
				}

				blockStr += "\n  }"
			}
			return blockStr, nil
		}

		// Otherwise, treat as a list attribute
		quotedItems := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			quotedItems[i] = fmt.Sprintf("%q", rv.Index(i).Interface())
		}
		return fmt.Sprintf("\n  %s = [%s]", key, strings.Join(quotedItems, ", ")), nil

	case reflect.Map:
		blockStr := fmt.Sprintf("\n  %s {", key)
		attributes, ok := value.(map[string]any)

		if !ok {
			return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
		}

		for k, v := range attributes {
			attr, err := formatTerraformAttribute(k, v)
			if err != nil {
				return "", err
			}
			blockStr += attr
		}

		blockStr += "\n  }"
		return blockStr, nil

	case reflect.Bool:
		return fmt.Sprintf("\n  %s = %t", key, value), nil

	case reflect.String:
		return fmt.Sprintf("\n  %s = %q", key, value), nil

	case reflect.Int:
		return fmt.Sprintf("\n  %s = %d", key, value), nil

	default:
		return fmt.Sprintf("\n  %s = %q", key, value), nil
	}
}
