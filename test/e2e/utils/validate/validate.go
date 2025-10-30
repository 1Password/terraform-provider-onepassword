package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfconfig "github.com/1Password/terraform-provider-onepassword/v2/test/e2e/terraform/config"
)

// buildItemChecks creates a list of test assertions to verify item attributes
func BuildItemChecks(resourceName string, attrs map[string]string) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "uuid"),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
	}

	for attr, expectedValue := range attrs {
		if attr == "tags" {
			checks = append(checks,
				resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "tags.0", expectedValue),
			)
			continue
		}

		checks = append(checks, resource.TestCheckResourceAttr(resourceName, attr, expectedValue))
	}

	return checks
}

// BuildPasswordRecipeChecks creates a list of test assertions to verify password recipe attributes with regex
func BuildPasswordRecipeChecks(resourceName string, recipe *tfconfig.PasswordRecipe) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{}

	if recipe.Length > 0 {
		checks = append(checks, checkPasswordPattern(resourceName, fmt.Sprintf("^.{%d}$", recipe.Length), "length"))
	}

	if recipe.Symbols {
		checks = append(checks, checkPasswordPattern(resourceName, `[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`+"`"+`]`, "symbols"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, `^[^!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~\`+"`"+`]+$`, "symbols"))
	}

	if recipe.Digits {
		checks = append(checks, checkPasswordPattern(resourceName, `[0-9]`, "digits"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, `^[^0-9]+$`, "digits"))
	}

	if recipe.Letters {
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
