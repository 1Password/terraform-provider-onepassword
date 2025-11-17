package password

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type PasswordRecipe struct {
	Length  int
	Symbols bool
	Digits  bool
}

// BuildPasswordRecipeChecks creates a list of test assertions to verify password recipe attributes
func BuildPasswordRecipeChecks(resourceName string, recipe PasswordRecipe) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "password_recipe.#", "1"),
	}

	length := recipe.Length
	symbols := recipe.Symbols
	digits := recipe.Digits

	// If length is not provided (0), the default is 32
	if recipe.Length == 0 {
		length = 32
	}

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

	return checks
}

func BuildPasswordRecipeMap(pr PasswordRecipe) map[string]any {
	return map[string]any{
		"length":  pr.Length,
		"symbols": pr.Symbols,
		"digits":  pr.Digits,
	}
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
