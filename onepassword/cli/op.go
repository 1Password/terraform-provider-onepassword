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

const (
	minimumOpCliVersion = "2.23.0" // introduction of stdin json support for `op item update`
)

type OP struct {
	binaryPath          string
	serviceAccountToken string
	account             string
}

func New(serviceAccountToken, binaryPath, account string) *OP {
	return &OP{
		binaryPath:          binaryPath,
		serviceAccountToken: serviceAccountToken,
		account:             account,
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

func (op *OP) checkCliVersion(ctx context.Context) error {
	cliVersion, err := op.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get version of op CLI: %w", err)
	}
	if cliVersion.LessThan(semver.MustParse(minimumOpCliVersion)) {
		return fmt.Errorf("current 1Password CLI version is \"%s\". Please upgrade to at least \"%s\"", cliVersion, minimumOpCliVersion)
	}
	return nil
}

func (op *OP) GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error) {
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return nil, versionErr
	}
	var res *onepassword.Vault
	err := op.execJson(ctx, &res, nil, p("vault"), p("get"), p(uuid))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error) {
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return nil, versionErr
	}
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
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return nil, versionErr
	}
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
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return nil, versionErr
	}
	return op.withRetry(func() (*onepassword.Item, error) {
		return op.create(ctx, item, vaultUuid)
	})
}

func (op *OP) create(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
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
	recipe := passwordRecipe(item)
	if recipe != "" {
		args = append(args, f("generate-password", recipe))
	}

	err = op.execJson(ctx, &res, payload, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) UpdateItem(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return nil, versionErr
	}
	return op.withRetry(func() (*onepassword.Item, error) {
		return op.update(ctx, item, vaultUuid)
	})
}

func (op *OP) update(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
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
	versionErr := op.checkCliVersion(ctx)
	if versionErr != nil {
		return versionErr
	}
	_, err := op.withRetry(func() (*onepassword.Item, error) {
		return op.delete(ctx, item, vaultUuid)
	})
	if err != nil {
		return err
	}
	return nil
}

func (op *OP) delete(ctx context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	if item.Vault.ID != "" && item.Vault.ID != vaultUuid {
		return nil, errors.New("item payload contains vault id that does not match vault uuid")
	}
	item.Vault.ID = vaultUuid

	return nil, op.execJson(ctx, nil, nil, p("item"), p("delete"), p(item.ID), f("vault", vaultUuid))
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

	if op.account != "" {
		args = append(args, f("account", op.account))
	}

	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg.format())
	}

	cmd := exec.CommandContext(ctx, op.binaryPath, cmdArgs...)
	cmd.Env = append(cmd.Environ(),
		"OP_FORMAT=json",
		//"OP_INTEGRATION_BUILDNUMBER="+version.ProviderVersion, // causes bad request errors from CLI
		"OP_INTEGRATION_NAME=terraform-provider",
		"OP_INTEGRATION_ID=TFP",
	)
	if op.serviceAccountToken != "" {
		cmd.Env = append(cmd.Env, "OP_SERVICE_ACCOUNT_TOKEN="+op.serviceAccountToken)
	}
	if op.account != "" {
		cmd.Env = append(cmd.Env, "OP_BIOMETRIC_UNLOCK_ENABLED=true")
	}

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

func (op *OP) withRetry(action func() (*onepassword.Item, error)) (*onepassword.Item, error) {
	attempt := 0
	item, err := action()
	if err != nil {
		// retry if there is 409 Conflict error
		if strings.Contains(err.Error(), "409") {
			// make 3 retry attempts to successfully finish the operation
			for attempt < 3 {
				waitBeforeRetry(attempt)
				item, err = action()
				if err != nil {
					attempt++
					continue
				}
				break
			}
		}
		// return error if operation did not succeed after retries
		// or error is other than 409
		if err != nil {
			return nil, err
		}
	}

	return item, nil
}
