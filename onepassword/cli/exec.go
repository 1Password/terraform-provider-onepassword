package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os/exec"
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

func (op *OP) exec(ctx context.Context, dst any, stdin []byte, args ...opArg) error {
	var cmdArgs []string
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg.format())
	}

	cmd := exec.CommandContext(ctx, op.binaryPath, cmdArgs...)
	cmd.Env = append(cmd.Environ(),
		"OP_SERVICE_ACCOUNT_TOKEN="+op.serviceAccountToken,
		"OP_FORMAT=json",
		"OP_INTEGRATION_NAME=terraform-provider-connect",
		"OP_INTEGRATION_ID=GO",
		//"OP_INTEGRATION_BUILDNUMBER="+version.ProviderVersion, // causes bad request errors from CLI
	)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	tflog.Debug(ctx, "running op command: "+cmd.String())

	result, err := cmd.Output()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return parseCliError(exitError.Stderr)
	}
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	if dst != nil {
		return json.Unmarshal(result, dst)
	}
	return nil
}
