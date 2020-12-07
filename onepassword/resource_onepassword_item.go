package onepassword

import (
	"fmt"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOnepasswordItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceOnepasswordItemCreate,
		Read:   resourceOnepasswordItemRead,
		Update: resourceOnepasswordItemUpdate,
		Delete: resourceOnepasswordItemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vault": {
				Type:     schema.TypeString,
				Required: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "login",
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"database": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
			},
			"password_recipe": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"length": {
							Type:        schema.TypeInt,
							Description: "The length of the password to be generated",
							Default:     32,
							Optional:    true,
							ForceNew:    true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(int)
								if v < 1 || v > 64 {
									errs = append(errs, fmt.Errorf("%q must be between 1 and 64 inclusive, got: %d", key, v))
								}
								return
							},
						},
						"letters": {
							Type:        schema.TypeBool,
							Default:     true,
							Description: "Should Letters [a-zA-Z] be used when generating passwords",
							ForceNew:    true,
							Optional:    true,
						},
						"digits": {
							Type:        schema.TypeBool,
							Default:     true,
							Description: "Should Letters [0-9] be used when generating passwords",
							ForceNew:    true,
							Optional:    true,
						},
						"symbols": {
							Type:        schema.TypeBool,
							Default:     true,
							Description: "Should special characters be used when generating passwords",
							ForceNew:    true,
							Optional:    true,
						},
					},
				},
				MaxItems: 1,
				Optional: true,
			},
		},
	}
}

func resourceOnepasswordItemCreate(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)
	vaultUUID := data.Get("vault").(string)

	item, err := dataToItem(data)
	if err != nil {
		return err
	}

	createdItem, err := client.CreateItem(item, vaultUUID)
	if err != nil {
		return err
	}

	itemToData(createdItem, data)

	return nil
}

func resourceOnepasswordItemRead(data *schema.ResourceData, meta interface{}) error {
	vaultUUID, itemUUID := vaultAndItemUUID(data.Id())

	client := meta.(connect.Client)
	item, err := client.GetItem(itemUUID, vaultUUID)

	if err != nil {
		return err
	}

	itemToData(item, data)

	return nil
}

func resourceOnepasswordItemUpdate(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)

	item, err := dataToItem(data)
	if err != nil {
		return err
	}

	updated, err := client.UpdateItem(item, data.Get("vault").(string))
	if err != nil {
		return err
	}

	itemToData(updated, data)

	return nil
}

func resourceOnepasswordItemDelete(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)

	item, err := dataToItem(data)
	if err != nil {
		return err
	}

	err = client.DeleteItem(item, data.Get("vault").(string))
	if err != nil {
		return err
	}

	return nil
}

func terraformID(item *onepassword.Item) string {
	return fmt.Sprintf("vaults/%s/items/%s", item.Vault.ID, item.ID)
}

func vaultAndItemUUID(tfID string) (vaultUUID, itemUUID string) {
	elements := strings.Split(tfID, "/")

	if len(elements) != 4 {
		return "", ""
	}

	return elements[1], elements[3]
}

func itemToData(item *onepassword.Item, data *schema.ResourceData) {
	data.SetId(terraformID(item))
	data.Set("uuid", item.ID)
	data.Set("vault", item.Vault.ID)
	data.Set("title", item.Title)

	for _, u := range item.URLs {
		if u.Primary {
			data.Set("url", u.URL)
		}
	}

	data.Set("tags", item.Tags)
	data.Set("category", strings.ToLower(string(item.Category)))

	for _, f := range item.Fields {
		switch f.Purpose {
		case "USERNAME":
			data.Set("username", f.Value)
		case "PASSWORD":
			data.Set("password", f.Value)
		default:
			data.Set(strings.ToLower(f.Label), f.Value)
		}
	}
}

func dataToItem(data *schema.ResourceData) (*onepassword.Item, error) {
	item := onepassword.Item{
		ID: data.Get("uuid").(string),
		Vault: onepassword.ItemVault{
			ID: data.Get("vault").(string),
		},
		Title: data.Get("title").(string),
		URLs: []onepassword.ItemURL{
			{
				Primary: true,
				URL:     data.Get("url").(string),
			},
		},
	}

	password := data.Get("password").(string)
	recipe, err := parseGeneratorRecipe(data.Get("password_recipe").([]interface{}))
	if err != nil {
		return nil, err
	}

	switch data.Get("category").(string) {
	case "login":
		item.Category = onepassword.Login
		item.Fields = []*onepassword.ItemField{
			{
				ID:      "username",
				Label:   "username",
				Purpose: "USERNAME",
				Type:    "STRING",
				Value:   data.Get("username").(string),
			},
			{
				ID:       "password",
				Label:    "password",
				Purpose:  "PASSWORD",
				Type:     "CONCEALED",
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
		}
	case "password":
		item.Category = onepassword.Password
		item.Fields = []*onepassword.ItemField{
			{
				ID:       "password",
				Label:    "password",
				Purpose:  "PASSWORD",
				Type:     "CONCEALED",
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
		}
	case "database":
		item.Category = onepassword.Database
		item.Fields = []*onepassword.ItemField{
			{
				ID:    "username",
				Label: "username",
				Type:  "STRING",
				Value: data.Get("username").(string),
			},
			{
				ID:       "password",
				Label:    "password",
				Type:     "CONCEALED",
				Value:    password,
				Generate: password == "",
				Recipe:   recipe,
			},
			{
				ID:    "hostname",
				Label: "hostname",
				Type:  "STRING",
				Value: data.Get("hostname").(string),
			},
			{
				ID:    "database",
				Label: "database",
				Type:  "STRING",
				Value: data.Get("database").(string),
			},
			{
				ID:    "port",
				Label: "port",
				Type:  "STRING",
				Value: data.Get("port").(string),
			},
			{
				ID:    "database_type",
				Label: "type",
				Type:  "MENU",
				Value: data.Get("type").(string),
			},
		}
	}

	return &item, nil
}

func parseGeneratorRecipe(recipe []interface{}) (*onepassword.GeneratorRecipe, error) {
	if recipe == nil || len(recipe) == 0 {
		return nil, nil
	}

	r := recipe[0].(map[string]interface{})

	parsed := &onepassword.GeneratorRecipe{
		Length:        32,
		CharacterSets: []string{},
	}

	length := r["length"].(int)
	if length > 64 {
		return nil, fmt.Errorf("password_recipe.length must be an integer between 1 and 64")
	}

	if length > 0 {
		parsed.Length = length
	}

	letters := r["letters"].(bool)
	if letters {
		parsed.CharacterSets = append(parsed.CharacterSets, "LETTERS")
	}

	digits := r["digits"].(bool)
	if digits {
		parsed.CharacterSets = append(parsed.CharacterSets, "DIGITS")
	}

	symbols := r["symbols"].(bool)
	if symbols {
		parsed.CharacterSets = append(parsed.CharacterSets, "SYMBOLS")
	}

	return parsed, nil
}
