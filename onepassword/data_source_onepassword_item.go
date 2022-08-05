package onepassword

import (
	"errors"
	"fmt"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOnepasswordItem() *schema.Resource {
	exactlyOneOfUUIDAndTitle := []string{"uuid", "title"}

	return &schema.Resource{
		Description: "Use this data source to get details of an item by its vault uuid and either the title or the uuid of the item.",
		Read:        dataSourceOnepasswordItemRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The Terraform resource identifier for this item in the format `vaults/<vault_id>/items/<item_id>`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault": {
				Description: vaultUUIDDescription,
				Type:        schema.TypeString,
				Required:    true,
			},
			"uuid": {
				Description:  "The UUID of the item to retrieve. This field will be populated with the UUID of the item if the item it looked up by its title.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: exactlyOneOfUUIDAndTitle,
			},
			"title": {
				Description:  "The title of the item to retrieve. This field will be populated with the title of the item if the item it looked up by its UUID.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: exactlyOneOfUUIDAndTitle,
			},
			"category": {
				Description: fmt.Sprintf(enumDescription, categoryDescription, categories),
				Type:        schema.TypeString,
				Computed:    true,
			},
			"url": {
				Description: urlDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hostname": {
				Description: dbHostnameDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"database": {
				Description: dbDatabaseDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				Description: dbPortDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: fmt.Sprintf(enumDescription, dbTypeDescription, dbTypes),
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tags": {
				Description: tagsDescription,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"username": {
				Description: usernameDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"password": {
				Description: passwordDescription,
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"note_value": {
				Description: noteValueDescription,
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Sensitive:   true,
			},
			"section": {
				Description: sectionsDescription,
				Type:        schema.TypeList,
				Computed:    true,
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
							Computed:    true,
						},
						"field": {
							Description: sectionFieldsDescription,
							Type:        schema.TypeList,
							Computed:    true,
							MinItems:    0,
							Elem: &schema.Resource{
								Description: fieldDescription,
								Schema: map[string]*schema.Schema{
									"id": {
										Description: fieldIDDescription,
										Type:        schema.TypeString,
										Computed:    true,
									},
									"label": {
										Description: fieldLabelDescription,
										Type:        schema.TypeString,
										Computed:    true,
									},
									"purpose": {
										Description: fmt.Sprintf(enumDescription, fieldPurposeDescription, fieldPurposes),
										Type:        schema.TypeString,
										Computed:    true,
									},
									"type": {
										Description: fmt.Sprintf(enumDescription, fieldTypeDescription, fieldTypes),
										Type:        schema.TypeString,
										Computed:    true,
									},
									"value": {
										Description: fieldValueDescription,
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceOnepasswordItemRead(data *schema.ResourceData, meta interface{}) error {
	client := meta.(connect.Client)

	item, err := getItemForDataSource(client, data)
	if err != nil {
		return err
	}

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

	dataSections := []interface{}{}
	for _, s := range item.Sections {
		section := map[string]interface{}{}

		section["id"] = s.ID
		section["label"] = s.Label

		fields := []interface{}{}

		for _, f := range item.Fields {
			if f.Section != nil && f.Section.ID == s.ID {
				dataField := map[string]interface{}{}
				dataField["id"] = f.ID
				dataField["label"] = strings.ToLower(f.Label)
				dataField["purpose"] = f.Purpose
				dataField["type"] = f.Type
				dataField["value"] = f.Value

				fields = append(fields, dataField)
			}
		}
		section["field"] = fields

		dataSections = append(dataSections, section)
	}

	data.Set("section", dataSections)

	for _, f := range item.Fields {
		switch f.Purpose {
		case "USERNAME":
			data.Set("username", f.Value)
		case "PASSWORD":
			data.Set("password", f.Value)
		case "NOTES":
			data.Set("note_value", f.Value)
		default:
			if f.Section == nil {
				data.Set(strings.ToLower(f.Label), f.Value)
			}
		}
	}

	return nil
}

func getItemForDataSource(client connect.Client, data *schema.ResourceData) (*onepassword.Item, error) {
	vaultUUID := data.Get("vault").(string)
	itemTitle := data.Get("title").(string)
	itemUUID := data.Get("uuid").(string)

	if itemTitle != "" {
		return client.GetItemByTitle(itemTitle, vaultUUID)
	}
	if itemUUID != "" {
		return client.GetItem(itemUUID, vaultUUID)
	}
	return nil, errors.New("uuid or title must be set")
}
