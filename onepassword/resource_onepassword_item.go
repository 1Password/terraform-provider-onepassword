package onepassword

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	itemUUIDDescription  = "The UUID of the item. Item identifiers are unique within a specific vault."
	vaultUUIDDescription = "The UUID of the vault the item is in."
	categoryDescription  = "The category of the item."
	itemTitleDescription = "The title of the item."
	urlDescription       = "The primary URL for the item."
	tagsDescription      = "An array of strings of the tags assigned to the item."
	usernameDescription  = "Username for this item."
	passwordDescription  = "Password for this item."

	dbHostnameDescription = "(Only applies to the database category) The address where the database can be found"
	dbDatabaseDescription = "(Only applies to the database category) The name of the database."
	dbPortDescription     = "(Only applies to the database category) The port the database is listening on."
	dbTypeDescription     = "(Only applies to the database category) The type of database."

	sectionsDescription      = "A list of custom sections in an item"
	sectionDescription       = "A custom section in an item that contains custom fields"
	sectionIDDescription     = "A unique identifier for the section."
	sectionLabelDescription  = "The label for the section."
	sectionFieldsDescription = "A list of custom fields in the section."

	fieldDescription        = "A custom field."
	fieldIDDescription      = "A unique identifier for the field."
	fieldLabelDescription   = "The label for the field."
	fieldPurposeDescription = "Purpose indicates this is a special field: a username, password, or notes field."
	fieldTypeDescription    = "The type of value stored in the field."
	fieldValueDescription   = "The value of the field."

	passwordRecipeDescription  = "The recipe used to generate a new value for a password."
	passwordElementDescription = "The kinds of characters to include in the password."
	passwordLengthDescription  = "The length of the password to be generated."
	passwordLettersDescription = "Use letters [a-zA-Z] when generating the password."
	passwordDigitsDescription  = "Use digits [0-9] when generating the password."
	passwordSymbolsDescription = "Use symbols [!@.-_*] when generating the password."

	enumDescription = "%s One of %q"
)

var categories = []string{"login", "password", "database"}
var dbTypes = []string{"db2", "filemaker", "msaccess", "mssql", "mysql", "oracle", "postgresql", "sqlite", "other"}
var fieldPurposes = []string{"USERNAME", "PASSWORD", "NOTES"}
var fieldTypes = []string{"STRING", "EMAIL", "CONCEALED", "URL", "TOTP", "DATE", "MONTH_YEAR", "MENU"}

func resourceOnepasswordItem() *schema.Resource {
	exactlyOneOfValueAndPasswordRecipe := []string{"value", "password_recipe"}

	return &schema.Resource{
		Description: "A 1Password item.",
		Create:      resourceOnepasswordItemCreate,
		Read:        resourceOnepasswordItemRead,
		Update:      resourceOnepasswordItemUpdate,
		Delete:      resourceOnepasswordItemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The Terraform resource identifier for this item in the format `vaults/<vault_id>/items/<item_id>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"uuid": {
				Description: itemUUIDDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault": {
				Description: vaultUUIDDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"category": {
				Description:  fmt.Sprintf(enumDescription, categoryDescription, categories),
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "login",
				ValidateFunc: validation.StringInSlice(categories, true),
				ForceNew:     true,
			},
			"title": {
				Description: itemTitleDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url": {
				Description: urlDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hostname": {
				Description: dbHostnameDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"database": {
				Description: dbDatabaseDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"port": {
				Description: dbPortDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  fmt.Sprintf(enumDescription, dbTypeDescription, dbTypes),
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dbTypes, true),
			},
			"tags": {
				Description: tagsDescription,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"username": {
				Description: usernameDescription,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"password": {
				Description:   passwordDescription,
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Computed:      true,
				ConflictsWith: []string{"password_recipe"},
			},
			"section": {
				Description: sectionsDescription,
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    0,
				Elem: &schema.Resource{
					Description: sectionDescription,
					Schema: map[string]*schema.Schema{
						"id": {
							Description: sectionIDDescription,
							Type:        schema.TypeString,
							Computed:    true,
						},
						"label": {
							Description: sectionLabelDescription,
							Type:        schema.TypeString,
							Required:    true,
						},
						"field": {
							Description: sectionFieldsDescription,
							Type:        schema.TypeList,
							Optional:    true,
							MinItems:    0,
							Elem: &schema.Resource{
								Description: fieldDescription,
								Schema: map[string]*schema.Schema{
									"id": {
										Description: fieldIDDescription,
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"label": {
										Description: fieldLabelDescription,
										Type:        schema.TypeString,
										Required:    true,
									},
									"purpose": {
										Description:  fmt.Sprintf(enumDescription, fieldPurposeDescription, fieldPurposes),
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(fieldPurposes, true),
									},
									"type": {
										Description:  fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
										Type:         schema.TypeString,
										Default:      "STRING",
										Optional:     true,
										ValidateFunc: validation.StringInSlice(fieldTypes, true),
									},
									"value": {
										Description:  fieldValueDescription,
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										Sensitive:    true,
										ExactlyOneOf: exactlyOneOfValueAndPasswordRecipe,
									},
									"password_recipe": passwordRecipe(func(s *schema.Schema) {
										s.ExactlyOneOf = exactlyOneOfValueAndPasswordRecipe
									}),
								},
							},
						},
					},
				},
			},
			"password_recipe": passwordRecipe(func(s *schema.Schema) {
				s.ConflictsWith = []string{
					"password",
				}
			}),
		},
	}
}

func passwordRecipe(modifier func(*schema.Schema)) *schema.Schema {
	s := &schema.Schema{
		Description: passwordDescription,
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Description: passwordElementDescription,
			Schema: map[string]*schema.Schema{
				"length": {
					Type:        schema.TypeInt,
					Description: passwordLengthDescription,
					Default:     32,
					Optional:    true,
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
					Description: passwordLettersDescription,
					Optional:    true,
				},
				"digits": {
					Type:        schema.TypeBool,
					Default:     true,
					Description: passwordDigitsDescription,
					Optional:    true,
				},
				"symbols": {
					Type:        schema.TypeBool,
					Default:     true,
					Description: passwordSymbolsDescription,
					Optional:    true,
				},
			},
		},
		MaxItems: 1,
		Optional: true,
	}
	modifier(s)
	return s
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

	dataSections := data.Get("section").([]interface{})
	for _, s := range item.Sections {
		section := map[string]interface{}{}
		newSection := true

		// Check for existing section state
		for i := 0; i < len(dataSections); i++ {
			existingSection := dataSections[i].(map[string]interface{})
			existingID := existingSection["id"].(string)
			existingLabel := existingSection["label"].(string)

			if (s.ID != "" && s.ID == existingID) || s.Label == existingLabel {
				section = existingSection
				newSection = false
			}
		}

		section["id"] = s.ID
		section["label"] = s.Label

		existingFields := []interface{}{}
		if section["field"] != nil {
			existingFields = section["field"].([]interface{})
		}
		for _, f := range item.Fields {
			if f.Section != nil && f.Section.ID == s.ID {
				dataField := map[string]interface{}{}
				newField := true
				// Check for existing field state
				for i := 0; i < len(existingFields); i++ {
					existingField := existingFields[i].(map[string]interface{})
					existingID := existingField["id"].(string)
					existingLabel := existingField["label"].(string)

					if (f.ID != "" && f.ID == existingID) || f.Label == existingLabel {
						dataField = existingFields[i].(map[string]interface{})
						newField = false
					}
				}

				dataField["id"] = f.ID
				dataField["label"] = f.Label
				dataField["purpose"] = f.Purpose
				dataField["type"] = f.Type
				dataField["value"] = f.Value

				if f.Recipe != nil {
					charSets := map[string]bool{}
					for _, s := range f.Recipe.CharacterSets {
						charSets[strings.ToLower(s)] = true
					}

					dataRecipe := map[string]interface{}{
						"length":  f.Recipe.Length,
						"letters": charSets["letters"],
						"digits":  charSets["digits"],
						"symbols": charSets["symbols"],
					}
					dataField["password_recipe"] = dataRecipe
				}

				if newField {
					existingFields = append(existingFields, dataField)
				}
			}
		}
		section["field"] = existingFields

		if newSection {
			dataSections = append(dataSections, section)
		}
	}

	data.Set("section", dataSections)

	for _, f := range item.Fields {
		switch f.Purpose {
		case "USERNAME":
			data.Set("username", f.Value)
		case "PASSWORD":
			data.Set("password", f.Value)
		default:
			if f.Section == nil {
				data.Set(f.Label, f.Value)
			}
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
		Tags: getTags(data),
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

	sections := data.Get("section").([]interface{})
	for i := 0; i < len(sections); i++ {
		section, ok := sections[i].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Unable to parse section: %v", sections[i])
		}
		sid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, fmt.Errorf("Unable to generate a section id: %w", err)
		}

		if section["id"].(string) != "" {
			sid = section["id"].(string)
		} else {
			section["id"] = sid
		}

		s := &onepassword.ItemSection{
			ID:    sid,
			Label: section["label"].(string),
		}
		item.Sections = append(item.Sections, s)

		sectionFields := section["field"].([]interface{})
		for j := 0; j < len(sectionFields); j++ {
			field, ok := sectionFields[j].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Unable to parse section field: %v", sectionFields[j])
			}

			f := &onepassword.ItemField{
				Section: s,
				ID:      field["id"].(string),
				Type:    field["type"].(string),
				Purpose: field["purpose"].(string),
				Label:   field["label"].(string),
				Value:   field["value"].(string),
			}

			recipe, err := parseGeneratorRecipe(field["password_recipe"].([]interface{}))
			if err != nil {
				return nil, err
			}

			if recipe != nil {
				addRecipe(f, recipe)
			}

			item.Fields = append(item.Fields, f)
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

func addRecipe(f *onepassword.ItemField, r *onepassword.GeneratorRecipe) {
	f.Recipe = r

	// Check to see if the current value adheres to the recipe

	var recipeLetters, recipeDigits, recipeSymbols bool
	hasLetters, _ := regexp.MatchString("[a-zA-Z]", f.Value)
	hasDigits, _ := regexp.MatchString("[0-9]", f.Value)
	hasSymbols, _ := regexp.MatchString("[^a-zA-Z0-9]", f.Value)

	for _, s := range r.CharacterSets {
		switch s {
		case "LETTERS":
			recipeLetters = true
		case "DIGITS":
			recipeDigits = true
		case "SYMBOLS":
			recipeSymbols = true
		}
	}

	if hasLetters != recipeLetters ||
		hasDigits != recipeDigits ||
		hasSymbols != recipeSymbols ||
		len(f.Value) != r.Length {
		f.Generate = true
	}
}

func getTags(data *schema.ResourceData) []string {
	tagInterface := data.Get("tags").([]interface{})
	tags := make([]string, len(tagInterface))
	for i, tag := range tagInterface {
		tags[i] = tag.(string)
	}
	return tags
}
