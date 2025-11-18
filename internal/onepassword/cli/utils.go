package cli

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

type opArg interface {
	format() string
}

type opFlag struct {
	name  string
	value string
}

func (f opFlag) format() string {
	return fmt.Sprintf("--%s=%s", f.name, f.value)
}

func f(name, value string) opArg {
	return opFlag{name: name, value: value}
}

type opParam struct {
	value string
}

func (p opParam) format() string {
	return p.value
}

func p(value string) opArg {
	return opParam{value: value}
}

var cliErrorRegex = regexp.MustCompile(`(?m)^\[ERROR] (\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) (.+)$`)

func parseCliError(stderr []byte) error {
	subMatches := cliErrorRegex.FindStringSubmatch(string(stderr))
	if len(subMatches) != 3 {
		return fmt.Errorf("unkown op error: %s", string(stderr))
	}
	return fmt.Errorf("op error: %s", subMatches[2])
}

func passwordField(item *model.Item) *model.ItemField {
	for i, f := range item.Fields {
		if f.Purpose == onepassword.FieldPurposePassword {
			return &item.Fields[i]
		}
	}
	return nil
}

func passwordRecipe(item *model.Item) string {
	if pf := passwordField(item); pf != nil {
		return passwordRecipeToString(pf.Recipe, pf.Generate)
	}
	return ""
}

func passwordRecipeToString(recipe *model.GeneratorRecipe, shouldGenerate bool) string {
	str := ""
	if shouldGenerate && recipe != nil {
		str += strings.Join(recipe.CharacterSets, ",")
		if recipe.Length > 0 {
			if str == "" {
				str += fmt.Sprintf("%d", recipe.Length)
			} else {
				str += fmt.Sprintf(",%d", recipe.Length)
			}
		}
	}
	return str
}

// waitBeforeRetry waits some amount of time based on retryAttempt
// it implements 'exponential backoff with jitter' algorithm
func waitBeforeRetry(retryAttempts int) {
	randInt, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		randInt = big.NewInt(0)
	}
	randPercentage := float64(randInt.Int64()) / 100
	jitter := (1.0 + randPercentage) / 2

	exp := math.Pow(2, float64(retryAttempts))
	retryTimeMilliseconds := 100 + 500*exp*jitter
	wait := time.Duration(retryTimeMilliseconds) * time.Millisecond
	time.Sleep(wait)
}

func makeBuildVersion(version string) string {
	parts := strings.Split(strings.ReplaceAll(version, "-beta", ""), ".")
	buildVersion := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) == 1 {
			buildVersion += "0" + parts[i]
		} else {
			buildVersion += parts[i]
		}
	}
	if len(parts) != 3 {
		return buildVersion
	}
	return buildVersion + "01"
}
