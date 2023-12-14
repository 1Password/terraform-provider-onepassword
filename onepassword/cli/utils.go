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

func passwordField(item *onepassword.Item) *onepassword.ItemField {
	for _, f := range item.Fields {
		if f.Purpose == onepassword.FieldPurposePassword {
			return f
		}
	}
	return nil
}

func passwordRecipe(item *onepassword.Item) string {
	str := ""
	if pf := passwordField(item); pf != nil {
		if pf.Recipe != nil {
			return passwordRecipeToString(pf.Recipe)
		}
	}
	return str
}

func passwordRecipeToString(recipe *onepassword.GeneratorRecipe) string {
	str := ""
	if recipe != nil {
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
