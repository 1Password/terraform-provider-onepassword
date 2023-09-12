package cli

import (
	"fmt"
	"regexp"
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
