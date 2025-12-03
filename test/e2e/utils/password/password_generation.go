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

// BuildPasswordRecipeChecks creates a list of test assertions to verify password recipe attributes at the item level
func BuildPasswordRecipeChecks(resourceName string, recipe PasswordRecipe) []resource.TestCheckFunc {
	return buildPasswordRecipeChecks(resourceName, "", "password", recipe)
}

// BuildPasswordRecipeChecksForField creates a list of test assertions to verify password recipe attributes for a field in a section
func BuildPasswordRecipeChecksForField(resourceName string, fieldPath string, recipe PasswordRecipe) []resource.TestCheckFunc {
	recipeAttrPath := fmt.Sprintf("%s.password_recipe", fieldPath)
	passwordAttrPath := fmt.Sprintf("%s.value", fieldPath)
	return buildPasswordRecipeChecks(resourceName, recipeAttrPath, passwordAttrPath, recipe)
}

// buildPasswordRecipeChecks is the shared implementation for password recipe checks
func buildPasswordRecipeChecks(resourceName, attrPrefix, passwordAttr string, recipe PasswordRecipe) []resource.TestCheckFunc {
	var recipeCheckPath string
	if attrPrefix == "" {
		recipeCheckPath = "password_recipe.#"
	} else {
		recipeCheckPath = fmt.Sprintf("%s.#", attrPrefix)
	}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, recipeCheckPath, "1"),
	}

	length := recipe.Length
	symbols := recipe.Symbols
	digits := recipe.Digits

	// If length is not provided (0), the default is 32
	if recipe.Length == 0 {
		length = 32
	}

	if length > 0 {
		checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, fmt.Sprintf("^.{%d}$", length), "length"))
	}

	// Letters are always included and not configurable
	checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, `[a-zA-Z]`, "letters"))

	if symbols {
		checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, `[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`+"`"+`]`, "symbols"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, `^[^!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~\`+"`"+`]+$`, "symbols"))
	}

	if digits {
		checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, `[0-9]`, "digits"))
	} else {
		checks = append(checks, checkPasswordPattern(resourceName, passwordAttr, `^[^0-9]+$`, "digits"))
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
func checkPasswordPattern(resourceName, passwordAttr, pattern, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		password := rs.Primary.Attributes[passwordAttr]
		matched, _ := regexp.MatchString(pattern, password)
		if !matched {
			return fmt.Errorf("password at %s does not match expected pattern: %s", passwordAttr, description)
		}
		return nil
	}
}
