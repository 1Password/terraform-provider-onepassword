package connectctx

import (
	"context"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
)

type Wrapper struct {
	client connect.Client
}

func (w *Wrapper) GetVault(_ context.Context, uuid string) (*onepassword.Vault, error) {
	return w.client.GetVault(uuid)
}

func (w *Wrapper) GetVaultsByTitle(_ context.Context, title string) ([]onepassword.Vault, error) {
	return w.client.GetVaultsByTitle(title)
}

func (w *Wrapper) GetItem(_ context.Context, itemUuid, vaultUuid string) (*onepassword.Item, error) {
	return w.client.GetItem(itemUuid, vaultUuid)
}

func (w *Wrapper) GetItemByTitle(_ context.Context, title string, vaultUuid string) (*onepassword.Item, error) {
	return w.client.GetItemByTitle(title, vaultUuid)
}

func (w *Wrapper) CreateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return w.client.CreateItem(item, vaultUuid)
}

func (w *Wrapper) UpdateItem(_ context.Context, item *onepassword.Item, vaultUuid string) (*onepassword.Item, error) {
	return w.client.UpdateItem(item, vaultUuid)
}

func (w *Wrapper) DeleteItem(_ context.Context, item *onepassword.Item, vaultUuid string) error {
	return w.client.DeleteItem(item, vaultUuid)
}

func Wrap(client connect.Client) *Wrapper {
	return &Wrapper{client: client}
}
