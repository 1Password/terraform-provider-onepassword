package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
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

func (op *OP) GetVault(ctx context.Context, uuid string) (*onepassword.Vault, error) {
	var res *onepassword.Vault
	err := op.exec(ctx, &res, nil, p("vault"), p("get"), p(uuid))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (op *OP) GetVaultsByTitle(ctx context.Context, title string) ([]onepassword.Vault, error) {
	var allVaults []onepassword.Vault
	err := op.exec(ctx, &allVaults, nil, p("vault"), p("list"))
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
	err := op.exec(ctx, &res, nil, p("item"), p("get"), p(itemUuid), f("vault", vaultUuid))
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
	err = op.exec(ctx, &res, payload, p("item"), p("create"), p("-"))
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

	// op cli does not support easy updating of an item by passing in a json payload
	err := op.DeleteItem(ctx, item, vaultUuid)
	if err != nil {
		return nil, err
	}

	// reset fields for item creation
	item.ID = ""
	item.Version = 0
	item.LastEditedBy = ""
	item.CreatedAt = time.Time{}
	item.UpdatedAt = time.Time{}

	return op.CreateItem(ctx, item, vaultUuid)
}

func (op *OP) DeleteItem(ctx context.Context, item *onepassword.Item, vaultUuid string) error {
	if item.Vault.ID != "" && item.Vault.ID != vaultUuid {
		return errors.New("item payload contains vault id that does not match vault uuid")
	}
	item.Vault.ID = vaultUuid

	return op.exec(ctx, nil, nil, p("item"), p("delete"), p(item.ID), f("vault", vaultUuid))
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
