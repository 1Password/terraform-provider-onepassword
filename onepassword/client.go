package onepassword

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
)

type cliClient struct{}

type (
	// Item represents an item returned to the consumer
	Item struct {
		ID    string `json:"id"`
		Title string `json:"title"`

		URLs     []onepassword.ItemURL `json:"urls,omitempty"`
		Favorite bool                  `json:"favorite,omitempty"`
		Tags     []string              `json:"tags,omitempty"`
		Version  int                   `json:"version,omitempty"`
		Trashed  bool                  `json:"trashed,omitempty"`

		Vault    onepassword.ItemVault    `json:"vault"`
		Category onepassword.ItemCategory `json:"category,omitempty"`

		Sections []*onepassword.ItemSection `json:"sections,omitempty"`
		Fields   []*ItemField               `json:"fields,omitempty"`
		Files    []*onepassword.File        `json:"files,omitempty"`

		LastEditedBy string    `json:"last_edited_by,omitempty"`
		CreatedAt    time.Time `json:"created_at,omitempty"`
		UpdatedAt    time.Time `json:"updated_at,omitempty"`
	}
	ItemField struct {
		ID       string                   `json:"id"`
		Section  *onepassword.ItemSection `json:"section,omitempty"`
		Type     string                   `json:"type"`
		Purpose  string                   `json:"purpose,omitempty"`
		Label    string                   `json:"label,omitempty"`
		Value    string                   `json:"value,omitempty"`
		Generate bool                     `json:"generate,omitempty"`
		Recipe   *GeneratorRecipe         `json:"recipe,omitempty"`
		Entropy  float64                  `json:"entropy,omitempty"`
	}
	GeneratorRecipe struct {
		Length        int      `json:"length,omitempty"`
		CharacterSets []string `json:"character_sets,omitempty"`
	}
)

func NewClient() *cliClient {
	return &cliClient{}
}

func NewItemFields(itemFields []*ItemField) []*onepassword.ItemField {
	if itemFields == nil {
		return nil
	}
	ret := make([]*onepassword.ItemField, 0, len(itemFields))
	for _, itemField := range itemFields {
		ret = append(ret, NewItemField(itemField))
	}
	return ret
}

func NewItemField(itemField *ItemField) *onepassword.ItemField {
	if itemField == nil {
		return nil
	}
	return &onepassword.ItemField{
		ID:       itemField.ID,
		Section:  itemField.Section,
		Type:     itemField.Type,
		Purpose:  itemField.Purpose,
		Label:    itemField.Label,
		Value:    itemField.Value,
		Generate: itemField.Generate,
		Recipe:   NewGeneratorRecipe(itemField.Recipe),
		Entropy:  itemField.Entropy,
	}
}

func NewGeneratorRecipe(generatorRecipe *GeneratorRecipe) *onepassword.GeneratorRecipe {
	if generatorRecipe == nil {
		return nil
	}
	return &onepassword.GeneratorRecipe{
		Length:        generatorRecipe.Length,
		CharacterSets: generatorRecipe.CharacterSets,
	}
}

func NewItem(item *Item) *onepassword.Item {
	if item == nil {
		return nil
	}
	return &onepassword.Item{
		ID:           item.ID,
		Title:        item.Title,
		URLs:         item.URLs,
		Favorite:     item.Favorite,
		Tags:         item.Tags,
		Version:      item.Version,
		Trashed:      item.Trashed,
		Vault:        item.Vault,
		Category:     item.Category,
		Sections:     item.Sections,
		Fields:       NewItemFields(item.Fields),
		Files:        item.Files,
		LastEditedBy: item.LastEditedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func (c *cliClient) GetVaults() ([]onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetVault(uuid string) (*onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetVaultsByTitle(uuid string) ([]onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	var temp *Item
	if out, err := exec.Command(
		"op", "item", "get", uuid,
		"--vault", vaultUUID,
		"--format", "json",
	).Output(); err != nil {
		return nil, err
	} else if err := json.Unmarshal(out, &temp); err != nil {
		return nil, err
	} else {
		return NewItem(temp), nil
	}
}

func (c *cliClient) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error) {
	var temp *Item
	if out, err := exec.Command(
		"op", "item", "get", "\""+title+"\"",
		"--vault", vaultUUID,
		"--format", "json",
	).Output(); err != nil {
		return nil, err
	} else if err := json.Unmarshal(out, &temp); err != nil {
		return nil, err
	} else {
		return NewItem(temp), nil
	}
}

func (c *cliClient) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetFile(fileUUID string, itemUUID string, vaultUUID string) (*onepassword.File, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliClient) GetFileContent(file *onepassword.File) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
