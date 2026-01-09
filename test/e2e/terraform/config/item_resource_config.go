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
			attr, err := formatTerraformAttribute(key, value, 1)
			if err != nil {
				return fmt.Sprintf("ERROR: %v", err)
			}
			resourceStr += attr
		}

		resourceStr += "\n}"
		return resourceStr
	}
}

// mapAttributeKeys defines which attributes should use map syntax (= {}) instead of block syntax
var mapAttributeKeys = map[string]bool{
	"section_map": true,
	"field_map":   true,
}

func formatTerraformAttribute(key string, value any, indent int) (string, error) {
	rv := reflect.ValueOf(value)
	indentStr := strings.Repeat("  ", indent)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		// Handle slices of maps recursively (block syntax)
		if rv.Type().Elem().Kind() == reflect.Map {
			blockStr := ""

			for i := 0; i < rv.Len(); i++ {
				blockStr += fmt.Sprintf("\n%s%s {", indentStr, key)
				attributes, ok := rv.Index(i).Interface().(map[string]any)

				if !ok {
					return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
				}

				for k, v := range attributes {
					attr, err := formatTerraformAttribute(k, v, indent+1)
					if err != nil {
						return "", err
					}
					blockStr += attr
				}

				blockStr += fmt.Sprintf("\n%s}", indentStr)
			}
			return blockStr, nil
		}

		// Otherwise, treat as a list attribute
		quotedItems := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			quotedItems[i] = fmt.Sprintf("%q", rv.Index(i).Interface())
		}
		return fmt.Sprintf("\n%s%s = [%s]", indentStr, key, strings.Join(quotedItems, ", ")), nil

	case reflect.Map:
		// Check if this should use map assignment syntax (= {}) or block syntax
		if mapAttributeKeys[key] {
			return formatMapAttribute(key, value, indent)
		}

		// Use block syntax for other maps (e.g., password_recipe)
		blockStr := fmt.Sprintf("\n%s%s {", indentStr, key)
		attributes, ok := value.(map[string]any)

		if !ok {
			return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
		}

		for k, v := range attributes {
			attr, err := formatTerraformAttribute(k, v, indent+1)
			if err != nil {
				return "", err
			}
			blockStr += attr
		}

		blockStr += fmt.Sprintf("\n%s}", indentStr)
		return blockStr, nil

	case reflect.Bool:
		return fmt.Sprintf("\n%s%s = %t", indentStr, key, value), nil

	case reflect.String:
		return fmt.Sprintf("\n%s%s = %q", indentStr, key, value), nil

	case reflect.Int:
		return fmt.Sprintf("\n%s%s = %d", indentStr, key, value), nil

	default:
		return fmt.Sprintf("\n%s%s = %q", indentStr, key, value), nil
	}
}

// formatMapAttribute formats a map attribute using Terraform map syntax: key = { "key1" = value1 }
func formatMapAttribute(key string, value any, indent int) (string, error) {
	indentStr := strings.Repeat("  ", indent)
	innerIndentStr := strings.Repeat("  ", indent+1)

	attributes, ok := value.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
	}

	if len(attributes) == 0 {
		return fmt.Sprintf("\n%s%s = {}", indentStr, key), nil
	}

	mapStr := fmt.Sprintf("\n%s%s = {", indentStr, key)

	for k, v := range attributes {
		// Each key in the map uses "key" = { ... } syntax
		nestedMap, isMap := v.(map[string]any)
		if isMap {
			mapStr += fmt.Sprintf("\n%s%q = {", innerIndentStr, k)
			for nestedKey, nestedValue := range nestedMap {
				attr, err := formatNestedMapValue(nestedKey, nestedValue, indent+2)
				if err != nil {
					return "", err
				}
				mapStr += attr
			}
			mapStr += fmt.Sprintf("\n%s}", innerIndentStr)
		} else {
			mapStr += fmt.Sprintf("\n%s%q = %v", innerIndentStr, k, formatValue(v))
		}
	}

	mapStr += fmt.Sprintf("\n%s}", indentStr)
	return mapStr, nil
}

// formatNestedMapValue formats values inside a nested map
func formatNestedMapValue(key string, value any, indent int) (string, error) {
	indentStr := strings.Repeat("  ", indent)

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Map:
		// Check if this should use map syntax
		if mapAttributeKeys[key] {
			return formatMapAttribute(key, value, indent)
		}

		// Use block syntax for nested blocks like password_recipe
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid terraform config: attribute %q has unsupported type %T", key, value)
		}

		blockStr := fmt.Sprintf("\n%s%s = {", indentStr, key)
		for k, v := range nestedMap {
			blockStr += fmt.Sprintf("\n%s  %s = %s", indentStr, k, formatValue(v))
		}
		blockStr += fmt.Sprintf("\n%s}", indentStr)
		return blockStr, nil

	case reflect.Bool:
		return fmt.Sprintf("\n%s%s = %t", indentStr, key, value), nil

	case reflect.String:
		return fmt.Sprintf("\n%s%s = %q", indentStr, key, value), nil

	case reflect.Int:
		return fmt.Sprintf("\n%s%s = %d", indentStr, key, value), nil

	default:
		return fmt.Sprintf("\n%s%s = %s", indentStr, key, formatValue(value)), nil
	}
}

// formatValue formats a single value for Terraform config
func formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%q", v)
	}
}
