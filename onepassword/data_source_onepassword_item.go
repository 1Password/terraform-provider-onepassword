package onepassword

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/1Password/connect-sdk-go/connect"
)

func dataSourceOnepasswordItem() *schema.Resource {
	return &schema.Resource{
		Description: "Get the contents of a 1Password item from its Item and Vault UUID.",
		Read:        dataSourceOnepasswordItemRead,
		Schema: map[string]*schema.Schema{
			"uuid": {
				Description: itemUUIDDescription,
				Type:        schema.TypeString,
				Required:    true,
			},
			"vault": {
				Description: vaultUUIDDescription,
				Type:        schema.TypeString,
				Required:    true,
			},
			"category": {
				Description:  fmt.Sprintf(enumDescription, categoryDescription, categories),
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "login",
				ValidateFunc: validation.StringInSlice(categories, true),
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
				Description: passwordDescription,
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
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
							Optional:    true,
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
										Description: fieldValueDescription,
										Type:        schema.TypeString,
										Optional:    true,
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

	vaultUUID := data.Get("vault").(string)
	itemUUID := data.Get("uuid").(string)

	data.SetId("")

	item, err := client.GetItem(itemUUID, vaultUUID)

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
		default:
			if f.Section == nil {
				data.Set(strings.ToLower(f.Label), f.Value)
			}
		}
	}

	return nil
}
