---
page_title: "onepassword_item Data Source - terraform-provider-onepassword"
subcategory: ""
description: |-
  Get the contents of a 1Password item from its Item and Vault UUID.
---

# Data Source `onepassword_item`

Get the contents of a 1Password item from its Item and Vault UUID.

## Example Usage

```terraform
data "onepassword_item" "example" {
  vault = var.demo_vault
  uuid  = onepassword_item.demo_sections.uuid
}
```

## Schema

### Required

- **uuid** (String, Required) The UUID of the item. Item identifiers are unique within a specific vault.
- **vault** (String, Required) The UUID of the vault the item is in.

### Optional

- **category** (String, Optional) The category of the item.
- **database** (String, Optional) (Only applies to the database category) The name of the database.
- **hostname** (String, Optional) (Only applies to the database category) The address where the database can be found
- **id** (String, Optional) The ID of this resource.
- **password** (String, Optional) Password for this item.
- **port** (String, Optional) (Only applies to the database category) The port the database is listening on.
- **section** (Block List) A list of custom sections in an item (see [below for nested schema](#nestedblock--section))
- **tags** (List of String, Optional) An array of strings of the tags assigned to the item.
- **title** (String, Optional) The title of the item.
- **type** (String, Optional) (Only applies to the database category) The type of database.
- **url** (String, Optional) The primary URL for the item.
- **username** (String, Optional) Username for this item.

<a id="nestedblock--section"></a>
### Nested Schema for `section`

Required:

- **label** (String, Required) The label for the section.

Optional:

- **field** (Block List) A list of custom fields in the section. (see [below for nested schema](#nestedblock--section--field))
- **id** (String, Optional) A unique identifier for the section.

<a id="nestedblock--section--field"></a>
### Nested Schema for `section.field`

Required:

- **label** (String, Required) The label for the field.

Optional:

- **id** (String, Optional) A unique identifier for the field.
- **purpose** (String, Optional) Purpose indicates this is a special field: a username, password, or notes field.
- **type** (String, Optional) The type of value stored in the field.
- **value** (String, Optional) The value of the field.


