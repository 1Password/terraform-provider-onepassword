package onepassword

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/1Password/connect-sdk-go/connect"
)

func dataSourceOnepasswordItem() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOnepasswordItemRead,
		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:     schema.TypeString,
				Required: true,
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
			},
			"section": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"label": {
										Type:     schema.TypeString,
										Required: true,
									},
									"purpose": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeString,
										Default:  "STRING",
										Optional: true,
									},
									"value": {
										Type:      schema.TypeString,
										Optional:  true,
										Computed:  true,
										Sensitive: true,
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
