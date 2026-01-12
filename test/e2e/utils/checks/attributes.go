package checks

import (
	"fmt"
	"strconv"

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

func buildAttributeChecks(resourceName string, attrPath string, expectedValue any) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	switch v := expectedValue.(type) {
	case []string:
		// Handle string slices
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.#", attrPath), strconv.Itoa(len(v))))
		for i, val := range v {
			checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.%d", attrPath, i), val))
		}

	case []map[string]any:
		// Handle nested block lists recursively
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("%s.#", attrPath), strconv.Itoa(len(v))))
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

// BuildSectionMapFieldValueCheck creates a check for a specific field value in section_map
func BuildSectionMapFieldValueCheck(resourceName, sectionLabel, fieldLabel, expectedValue string) resource.TestCheckFunc {
	path := fmt.Sprintf("section_map.%s.field_map.%s.value", sectionLabel, fieldLabel)
	return resource.TestCheckResourceAttr(resourceName, path, expectedValue)
}

// BuildSectionMapFieldIDSetCheck creates a check that a field ID is set in section_map
func BuildSectionMapFieldIDSetCheck(resourceName, sectionLabel, fieldLabel string) resource.TestCheckFunc {
	path := fmt.Sprintf("section_map.%s.field_map.%s.id", sectionLabel, fieldLabel)
	return resource.TestCheckResourceAttrSet(resourceName, path)
}

// BuildSectionMapIDSetCheck creates a check that a section ID is set in section_map
func BuildSectionMapIDSetCheck(resourceName, sectionLabel string) resource.TestCheckFunc {
	path := fmt.Sprintf("section_map.%s.id", sectionLabel)
	return resource.TestCheckResourceAttrSet(resourceName, path)
}
