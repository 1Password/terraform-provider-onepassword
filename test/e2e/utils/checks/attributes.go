package checks

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// BuildItemChecks creates a list of test assertions to verify item attributes
func BuildItemChecks(resourceName string, attrs map[string]any) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "uuid"),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
	}

	for attr, expectedValue := range attrs {
		checks = append(checks, buildAttributeChecks(resourceName, attr, expectedValue)...)
	}

	return checks
}

func buildAttributeChecks(resourceName, attrPath string, expectedValue any) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	switch v := expectedValue.(type) {
	case []string:
		// Handle string slices
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.#", attrPath), fmt.Sprintf("%d", len(v))))
		for i, val := range v {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.%d", attrPath, i), val))
		}

	case []map[string]any:
		// Handle nested block lists recursively
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.#", attrPath), fmt.Sprintf("%d", len(v))))
		for i, nestedMap := range v {
			for nestedAttr, nestedValue := range nestedMap {
				nestedPath := fmt.Sprintf("%s.%d.%s", attrPath, i, nestedAttr)
				checks = append(checks, buildAttributeChecks(resourceName, nestedPath, nestedValue)...)
			}
		}

	default:
		// Handle simple attributes
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, attrPath, fmt.Sprintf("%v", expectedValue)))
	}

	return checks
}
