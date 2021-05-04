package onepassword

import (
	"fmt"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var items = map[string]*onepassword.Item{}

var schemaKeys = []string{
	"uuid",
	"vault",
	"category",
	"title",
	"url",
	"hostname",
	"database",
	"port",
	"username",
	"password",
	"tags",
}

func TestResourceOnepasswordLoginItemCRUD(t *testing.T) {
	itemToCreate := generateResourceLoginItem(t)
	testCRUDForItem(t, itemToCreate)
}

func TestResourceOnepasswordDatabaseItemCRUD(t *testing.T) {
	itemToCreate := generateResourceDatabaseItem(t)
	testCRUDForItem(t, itemToCreate)
}

func TestResourceOnepasswordPasswordItemCRUD(t *testing.T) {
	itemToCreate := generateResourcePasswordItem(t)
	testCRUDForItem(t, itemToCreate)
}

func TestAddSectionsToItem(t *testing.T) {
	item := generateResourceLoginItem(t)

	section1 := map[string]interface{}{
		"label": "Section One",
		"id":    "123",
		"field": []interface{}{
			map[string]interface{}{
				"label": "token",
				"type":  "CONCEALED",
				"value": "123",
			},
			map[string]interface{}{
				"label": "user",
				"value": "root",
			},
		},
	}

	section2 := map[string]interface{}{
		"label": "Section One",
		"id":    "123",
		"field": []interface{}{
			map[string]interface{}{
				"label": "secret value",
				"type":  "CONCEALED",
				"password_recipe": []interface{}{
					map[string]interface{}{
						"length": 42,
						"digits": false,
					},
				},
			},
		},
	}

	item.Set("section", []interface{}{section1, section2})

	testCRUDForItem(t, item)
}

func testCRUDForItem(t *testing.T, itemToCreate *schema.ResourceData) {
	meta := &testClient{}
	DoGetItemFunc = getItem
	DoCreateItemFunc = createItem
	DoDeleteItemFunc = deleteItem
	DoUpdateItemFunc = updateItem

	// Creating an Item
	err := resourceOnepasswordItemCreate(itemToCreate, meta)
	if err != nil {
		t.Errorf("Unexpected error occured when creating item")
	}

	storedItem := items[itemToCreate.Get("uuid").(string)]
	compareItemToSource(t, itemToCreate, storedItem)

	// Reading an Item
	itemRead := generateResource(t, storedItem)
	err = resourceOnepasswordItemRead(itemRead, meta)
	if err != nil {
		t.Errorf("Unexpected error occured when reading item")
	}
	compareResources(t, itemToCreate, itemRead)

	//Updating an item
	itemToCreate.Set("password", "new_password")
	err = resourceOnepasswordItemUpdate(itemToCreate, meta)
	if err != nil {
		t.Errorf("Unexpected error occured when deleting item")
	}
	err = resourceOnepasswordItemRead(itemRead, meta)
	if err != nil {
		t.Errorf("Unexpected error occured when reading item")
	}
	compareResources(t, itemToCreate, itemRead)

	//Deleting an item
	err = resourceOnepasswordItemDelete(itemRead, meta)
	if err != nil {
		t.Errorf("Unexpected error occured when deleting item")
	}

	// Reading an Item That No Longer Exists
	err = resourceOnepasswordItemRead(itemRead, meta)
	if err == nil {
		t.Errorf("Expected an error when retrieving a nonexistent item")
	}
}

func getItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	item, found := items[uuid]
	if !found {
		return nil, fmt.Errorf("Could not retrieve item with id %v", uuid)
	}
	return item, nil
}

func createItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	for i := 0; i < len(item.Fields); i++ {
		if item.Fields[i].Recipe != nil {
			item.Fields[i].Recipe = nil
			if item.Fields[i].Generate {
				item.Fields[i].Generate = false
				item.Fields[i].Value = "GENERATEDVALUE"
			}
		}
	}

	items[item.ID] = item
	return item, nil
}

func deleteItem(item *onepassword.Item, vaultUUID string) error {
	delete(items, item.ID)
	return nil
}

func updateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	for i := 0; i < len(item.Fields); i++ {
		if item.Fields[i].Recipe != nil {
			item.Fields[i].Recipe = nil
			if item.Fields[i].Generate {
				item.Fields[i].Generate = false
				item.Fields[i].Value = "GENERATEDVALUE"
			}
		}
	}

	items[item.ID] = item
	return item, nil
}

func generateResourceLoginItem(t *testing.T) *schema.ResourceData {
	resourceData := generateBaseItem(t)
	resourceData.Set("category", "login")
	resourceData.Set("username", "test_user")
	resourceData.Set("password", "test_password")
	return resourceData
}

func generateResourceDatabaseItem(t *testing.T) *schema.ResourceData {
	resourceData := generateResourceLoginItem(t)
	resourceData.Set("category", "database")
	resourceData.Set("hostname", "test_host")
	resourceData.Set("database", "test_database")
	resourceData.Set("port", "test_port")
	resourceData.Set("type", "test_type")
	return resourceData
}

func generateResourcePasswordItem(t *testing.T) *schema.ResourceData {
	resourceData := generateBaseItem(t)
	resourceData.Set("category", "password")
	resourceData.Set("password", "test_password")
	return resourceData
}

func generateBaseItem(t *testing.T) *schema.ResourceData {
	resourceData := schema.TestResourceDataRaw(t, resourceOnepasswordItem().Schema, nil)
	resourceData.Set("uuid", "79841a98-dd4a-4c34-8be5-32dca20a7328")
	resourceData.Set("vault", "df2e9643-45ad-4ff9-8b98-996f801afa75")
	resourceData.Set("title", "test_login")
	resourceData.Set("url", "some_url")
	resourceData.Set("tags", []string{"tag-1", "tag-2"})
	return resourceData
}

func generateResource(t *testing.T, item *onepassword.Item) *schema.ResourceData {
	resourceData := schema.TestResourceDataRaw(t, resourceOnepasswordItem().Schema, nil)
	resourceData.Set("uuid", item.ID)
	resourceData.Set("vault", item.Vault.ID)
	resourceData.SetId(fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID))
	return resourceData
}

func compareResources(t *testing.T, expectedResource, actualResource *schema.ResourceData) {
	t.Helper()
	for _, key := range schemaKeys {
		if key == "tags" {
			compareStringSlice(t, getTags(actualResource), getTags(expectedResource))
		} else {
			if expectedResource.Get(key) != actualResource.Get(key) {
				t.Errorf("Expected %v to be %v but was %v", key, actualResource.Get(key), expectedResource.Get(key))
			}
		}

	}
}
