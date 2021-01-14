package onepassword

import "github.com/1Password/connect-sdk-go/onepassword"

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

var (
	DoGetVaultsFunc        func() ([]onepassword.Vault, error)
	DoGetVaultsByTitleFunc func(title string) ([]onepassword.Vault, error)
	DoGetItemFunc          func(uuid string, vaultUUID string) (*onepassword.Item, error)
	DoGetItemsByTitleFunc  func(title string, vaultUUID string) ([]onepassword.Item, error)
	DoGetItemByTitleFunc   func(title string, vaultUUID string) (*onepassword.Item, error)
	DoCreateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
	DoDeleteItemFunc       func(item *onepassword.Item, vaultUUID string) error
	DoGetItemsFunc         func(vaultUUID string) ([]onepassword.Item, error)
	DoUpdateItemFunc       func(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
)

// Do is the mock client's `Do` func
func (m *testClient) GetVaults() ([]onepassword.Vault, error) {
	return DoGetVaultsFunc()
}

func (m *testClient) GetVaultsByTitle(title string) ([]onepassword.Vault, error) {
	return DoGetVaultsByTitleFunc(title)
}

func (m *testClient) GetItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	return DoGetItemFunc(uuid, vaultUUID)
}

func (m *testClient) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	return DoGetItemsFunc(vaultUUID)
}

func (m *testClient) GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error) {
	return DoGetItemsByTitleFunc(title, vaultUUID)
}

func (m *testClient) GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error) {
	return DoGetItemByTitleFunc(title, vaultUUID)
}

func (m *testClient) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return DoCreateItemFunc(item, vaultUUID)
}

func (m *testClient) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	return DoDeleteItemFunc(item, vaultUUID)
}

func (m *testClient) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return DoUpdateItemFunc(item, vaultUUID)
}
