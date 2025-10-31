package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func BuildPasswordRecipeChecks(resourceName string, recipe map[string]any) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "password_recipe.#", "1"),
	}

	length, _ := recipe["length"].(int)
	symbols, _ := recipe["symbols"].(bool)
	digits, _ := recipe["digits"].(bool)
	letters, _ := recipe["letters"].(bool)

	if length > 0 {
		checks = append(checks, checkPasswordPattern(resourceName, fmt.Sprintf("^.{%d}$", length), "length"))
	}

	if symbols {
		checks = append(checks, checkPasswordPattern(resourceName, `[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`+"`"+`]`, "symbols"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, `^[^!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~\`+"`"+`]+$`, "symbols"))
	}

	if digits {
		checks = append(checks, checkPasswordPattern(resourceName, `[0-9]`, "digits"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, `^[^0-9]+$`, "digits"))
	}

	if letters {
		checks = append(checks, checkPasswordPattern(resourceName, `[a-zA-Z]`, "letters"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, `^[^a-zA-Z]+$`, "letters"))
	}

	return checks
}

// checkPasswordPattern creates a test assertion to verify password pattern with regex
func checkPasswordPattern(resourceName, pattern, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		password := rs.Primary.Attributes["password"]
		matched, _ := regexp.MatchString(pattern, password)

		if !matched {
			return fmt.Errorf("password does not match expected pattern: %s", description)
		}
		return nil
	}
}
