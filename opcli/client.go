package opcli

import (
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
)

const (
	defaultOnePasswordPath = "/usr/local/bin/op"
)

type cliProvider struct {
	cli OnePasswordCLI
}

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

func NewCLIClient(account, password string) (*cliProvider, error) {
	cli, err := NewOnePasswordCLI(account, password)
	if err != nil {
		return nil, err
	}
	return &cliProvider{
		cli: cli,
	}, nil
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

func (c *cliProvider) GetVaults() ([]onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) GetVault(uuid string) (*onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) GetVaultsByTitle(uuid string) ([]onepassword.Vault, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) GetItem(uuid string, vaultUUID string) (*onepassword.Item, error) {
	return c.cli.GetItem(uuid, vaultUUID)
}

func (c *cliProvider) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	return []onepassword.Item{}, nil
}

func (c *cliProvider) GetItemsByTitle(title string, vaultUUID string) ([]onepassword.Item, error) {
	return []onepassword.Item{}, nil
}

func (c *cliProvider) GetItemByTitle(title string, vaultUUID string) (*onepassword.Item, error) {
	return c.cli.GetItem(title, vaultUUID)
}

func (c *cliProvider) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) GetFile(fileUUID string, itemUUID string, vaultUUID string) (*onepassword.File, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cliProvider) GetFileContent(file *onepassword.File) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func NewOnePasswordCLI(account, password string) (OnePasswordCLI, error) {
	token, err := getOnePasswordSessionToken(account, password)
	if err != nil {
		return OnePasswordCLI{}, err
	}

	return OnePasswordCLI{
		account: account,
		token:   token,
	}, nil
}
