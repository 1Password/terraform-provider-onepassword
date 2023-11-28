package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type OP struct {
	binaryPath          string
	serviceAccountToken string
}

func New(serviceAccountToken, binaryPath string) *OP {
	return &OP{
		binaryPath:          binaryPath,
		serviceAccountToken: serviceAccountToken,
	}
}

func (op *OP) GetVersion(ctx context.Context) (*semver.Version, error) {
	result, err := op.execRaw(ctx, nil, p("--version"))
	if err != nil {
		return nil, err
	}
	versionString := strings.TrimSpace(string(result))
	version, err := semver.NewVersion(versionString)
	if err != nil {
		return nil, fmt.Errorf("%w (input is: %s)", err, versionString)
	}
	return version, nil
}

func (op *OP) GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error) {
	var res *onepassword.Vault
	err := op.execJson(ctx, &res, nil, p("vault"), p("get"), p(uuid))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error) {
	var allVaults []onepassword.Vault
	err := op.execJson(ctx, &allVaults, nil, p("vault"), p("list"))
	if err != nil {
		return nil, err
	}

	var res []onepassword.Vault
	for _, v := range allVaults {
		if v.Name == title {
			res = append(res, v)
		}
	}
	return res, nil
}

func (op *OP) GetItem(ctx context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error) {
	var res *onepassword.Item
	err := op.execJson(ctx, &res, nil, p("item"), p("get"), p(itemUuid), f("vault", vaultUuid))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) GetItemByTitle(ctx context.Context, title string, vaultUuid string) (*onepassword.Item, error) {
	return op.GetItem(ctx, title, vaultUuid)
}

func (op *OP) CreateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	if item.Vault.ID != "" && item.Vault.ID != vaultUuid {
		return nil, errors.New("item payload contains vault id that does not match vault uuid")
	}
	item.Vault.ID = vaultUuid

	payload, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var res *onepassword.Item
	args := []opArg{p("item"), p("create"), p("-")}
	// 'op item create' command doesn't support generating passwords when using templates
	// therefore need to use --generate-password flag to set it
	if pf := passwordField(item); pf != nil {
		recipeStr := "letters,digits,32"
		if pf.Recipe != nil {
			recipeStr = passwordRecipeToString(pf.Recipe)
		}
		args = append(args, f("generate-password", recipeStr))
	}

	err = op.execJson(ctx, &res, payload, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	if item.Vault.ID != "" && item.Vault.ID != vaultUuid {
		return nil, errors.New("item payload contains vault id that does not match vault uuid")
	}
	item.Vault.ID = vaultUuid

	payload, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var res *onepassword.Item
	err = op.execJson(ctx, &res, payload, p("item"), p("edit"), p(item.ID), f("vault", vaultUuid))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error {
	if item.Vault.ID != "" && item.Vault.ID != vaultUuid {
		return errors.New("item payload contains vault id that does not match vault uuid")
	}
	item.Vault.ID = vaultUuid

	return op.execJson(ctx, nil, nil, p("item"), p("delete"), p(item.ID), f("vault", vaultUuid))
}

func (op *OP) execJson(ctx context.Context, dst any, stdin []byte, args ...opArg) error {
	result, err := op.execRaw(ctx, stdin, args...)
	if err != nil {
		return err
	}
	if dst != nil {
		return json.Unmarshal(result, dst)
	}
	return nil
}

func (op *OP) execRaw(ctx context.Context, stdin []byte, args ...opArg) ([]byte, error) {
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
		return nil, parseCliError(exitError.Stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return result, nil
}
