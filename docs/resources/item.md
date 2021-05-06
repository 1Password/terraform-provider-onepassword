---
page_title: "onepassword_item Resource - terraform-provider-onepassword"
subcategory: ""
description: |-
  A 1Password item.
---

# Resource `onepassword_item`

A 1Password item.

## Example Usage

```terraform
resource "onepassword_item" "demo_password" {
  vault = var.demo_vault

  title    = "Demo Password Recipe"
  category = "password"

  password_recipe {
    length  = 40
    symbols = false
  }
}

resource "onepassword_item" "demo_login" {
  vault = var.demo_vault

  title    = "Demo Terraform Login"
  category = "login"
  username = "test@example.com"
}

resource "onepassword_item" "demo_db" {
  vault    = var.demo_vault
  category = "database"
  type     = "mysql"

  title    = "Demo TF Database"
  username = "root"

  database = "Example MySQL Instance"
  hostname = "localhost"
  port     = 3306
}
```

## Schema

### Required

- **vault** (String, Required) The UUID of the vault the item is in.

### Optional

- **category** (String, Optional) The category of the item.
- **database** (String, Optional) (Only applies to the database category) The name of the database.
- **hostname** (String, Optional) (Only applies to the database category) Where can the database be found.
- **id** (String, Optional) The ID of this resource.
- **password** (String, Optional) Password for this item.
- **password_recipe** (Block List, Max: 1) Password for this item. (see [below for nested schema](#nestedblock--password_recipe))
- **port** (String, Optional) (Only applies to the database category) The port is the database listening on.
- **section** (Block List) A list of custom sections on an item (see [below for nested schema](#nestedblock--section))
- **tags** (List of String, Optional) An array of strings of the tags assigned to the item.
- **title** (String, Optional) The title of the item.
- **type** (String, Optional) (Only applies to the database category) The type of database.
- **url** (String, Optional) The primary URL where the resource is applicable.
- **username** (String, Optional) Username for this item.
- **uuid** (String, Optional) The UUID of the item. Item identifiers are unique within a specific Vault.

<a id="nestedblock--password_recipe"></a>
### Nested Schema for `password_recipe`

Optional:

- **digits** (Boolean, Optional) Should Letters [0-9] be used when generating passwords
- **length** (Number, Optional) The length of the password to be generated
- **letters** (Boolean, Optional) Should Letters [a-zA-Z] be used when generating passwords
- **symbols** (Boolean, Optional) Should special characters be used when generating passwords


<a id="nestedblock--section"></a>
### Nested Schema for `section`

Required:

- **label** (String, Required) The label for the section.

Optional:

- **field** (Block List) A list of custom fields in this section. (see [below for nested schema](#nestedblock--section--field))
- **id** (String, Optional) A unique identifier for the section.

<a id="nestedblock--section--field"></a>
### Nested Schema for `section.field`

Required:

- **label** (String, Required) The label for the field.

Optional:

- **id** (String, Optional) A unique identifier for the field.
- **password_recipe** (Block List, Max: 1) Password for this item. (see [below for nested schema](#nestedblock--section--field--password_recipe))
- **purpose** (String, Optional) Purpose indicates this is the special username, password, or notes field
- **type** (String, Optional) The type of value stored in the field.
- **value** (String, Optional) The value of the field

<a id="nestedblock--section--field--password_recipe"></a>
### Nested Schema for `section.field.password_recipe`

Optional:

- **digits** (Boolean, Optional) Should Letters [0-9] be used when generating passwords
- **length** (Number, Optional) The length of the password to be generated
- **letters** (Boolean, Optional) Should Letters [a-zA-Z] be used when generating passwords
- **symbols** (Boolean, Optional) Should special characters be used when generating passwords

## Import

Import is supported using the following syntax:

```shell
# import an existing 1Password item
terraform import onepassword_item.myitem vaults/<vault uuid>/items/<item uuid>
```
