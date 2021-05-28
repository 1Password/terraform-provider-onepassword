package onepassword

import (
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
)

type testClient struct {
	GetVaultsFunc        func() ([]onepassword.Vault, error)
	GetVaultsByTitleFunc func(title string) ([]onepassword.Vault, error)
	GetItemFunc          func(uuid string, vaultUUID string) (*onepassword.Item, error)
	GetItemsFunc         func(vaultUUID string) ([]onepassword.Item, error)
	GetItemsByTitleFunc  func(title string, vaultUUID string) ([]onepassword.Item, error)
	GetItemByTitleFunc   func(title string, vaultUUID string) (*onepassword.Item, error)
	CreateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	UpdateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	DeleteItemFunc       func(item *onepassword.Item, vaultUUID string) error
}

var _ connect.Client = (*testClient)(nil)

// Do is the mock client's `Do` func
func (m *testClient) GetVaults() ([]onepassword.Vault, error) {
	return m.GetVaultsFunc()
}

func (m *testClient) GetVaultsByTitle(title string) ([]onepassword.Vault, error) {
	return m.GetVaultsByTitleFunc(title)
}

func (m *testClient) GetItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	return m.GetItemFunc(uuid, vaultUUID)
}

func (m *testClient) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	return m.GetItemsFunc(vaultUUID)
}

func (m *testClient) GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error) {
	return m.GetItemsByTitleFunc(title, vaultUUID)
}

func (m *testClient) GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error) {
	return m.GetItemByTitleFunc(title, vaultUUID)
}

func (m *testClient) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return m.CreateItemFunc(item, vaultUUID)
}

func (m *testClient) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	return m.DeleteItemFunc(item, vaultUUID)
}

func (m *testClient) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return m.UpdateItemFunc(item, vaultUUID)
}
