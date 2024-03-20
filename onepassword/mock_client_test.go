package onepassword

import (
	"context"

	"github.com/1Password/connect-sdk-go/onepassword"
)

type testClient struct {
	GetVaultFunc         func(vaultUUID string) (*onepassword.Vault, error)
	GetVaultsByTitleFunc func(title string) ([]onepassword.Vault, error)
	GetItemFunc          func(uuid string, vaultUUID string) (*onepassword.Item, error)
	GetItemByTitleFunc   func(title string, vaultUUID string) (*onepassword.Item, error)
	CreateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	UpdateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	DeleteItemFunc       func(item *onepassword.Item, vaultUUID string) error
	GetFileFunc          func(file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error)
}

var _ Client = (*testClient)(nil)

func (m *testClient) GetVault(_ context.Context, vaultUUID string) (*onepassword.Vault, error) {
	return m.GetVaultFunc(vaultUUID)
}

func (m *testClient) GetVaultsByTitle(_ context.Context, title string) ([]onepassword.Vault, error) {
	return m.GetVaultsByTitleFunc(title)
}

func (m *testClient) GetItem(_ context.Context, uuid string, vaultUUID string) (*onepassword.Item, error) {
	return m.GetItemFunc(uuid, vaultUUID)
}

func (m *testClient) GetItemByTitle(_ context.Context, title string, vaultUUID string) (*onepassword.Item, error) {
	return m.GetItemByTitleFunc(title, vaultUUID)
}

func (m *testClient) CreateItem(_ context.Context, item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return m.CreateItemFunc(item, vaultUUID)
}

func (m *testClient) DeleteItem(_ context.Context, item *onepassword.Item, vaultUUID string) error {
	return m.DeleteItemFunc(item, vaultUUID)
}

func (m *testClient) UpdateItem(_ context.Context, item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return m.UpdateItemFunc(item, vaultUUID)
}

func (m *testClient) GetFileContent(_ context.Context, file *onepassword.File, itemUUID, vaultUUID string) ([]byte, error) {
	return m.GetFileFunc(file, itemUUID, vaultUUID)
}
