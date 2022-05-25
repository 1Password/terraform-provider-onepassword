package opcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/1Password/connect-sdk-go/onepassword"
)

type OnePasswordCLI struct {
	account string
	token   string
}

func (cli OnePasswordCLI) GetItem(item, vault string) (*onepassword.Item, error) {
	output, err := cli.command(
		defaultOnePasswordPath,
		"item", "get", item,
		"--vault", vault,
		"--format", "json",
	)
	if err != nil {
		return nil, err
	}

	var value *onepassword.Item
	err = json.Unmarshal(output, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshlal: %v data: %v", err, string(output))
	}
	return value, nil
}

func (cli OnePasswordCLI) command(name string, args ...string) (output []byte, err error) {
	if cli.token == "" {
		return nil, errors.New("OP client not authenticated")
	}
	args = append(args,
		"--session", cli.token,
		"--account", cli.account,
		"--no-color",
	)
	cmd := exec.Command(name, args...)
	var stdout, stdin, stderr bytes.Buffer

	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stdout.Write(stderr.Bytes())
		return nil, fmt.Errorf(string(stdout.Bytes()))
	}
	return stdout.Bytes(), nil
}

func getOnePasswordSessionToken(account, password string) (string, error) {
	cmd := exec.Command(defaultOnePasswordPath,
		"signin",
		"--account", account,
		"--raw",
	)
	var stdout, stdin, stderr bytes.Buffer
	cmd.Env = os.Environ()

	stdin.Write([]byte(password))
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stdout.Write(stderr.Bytes())
		return "", fmt.Errorf("failed to sign in: %v", string(stdout.Bytes()))
	}

	return string(stdout.Bytes()), nil
}
