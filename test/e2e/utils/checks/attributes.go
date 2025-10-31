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
		// Check if the value is a slice and iterate over it
		if slice, ok := expectedValue.([]string); ok {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.#", attr), fmt.Sprintf("%d", len(slice))))

			for i, val := range slice {
				checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.%d", attr, i), val))
			}
		} else {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, attr, fmt.Sprintf("%v", expectedValue)))
		}
	}

	return checks
}
