---
page_title: "onepassword_vault Data Source - terraform-provider-onepassword"
subcategory: ""
description: |-
  Use this data source to get details of a vault by either its name or uuid.
---

# Data Source `onepassword_vault`

Use this data source to get details of a vault by either its name or uuid.



## Schema

### Optional

- **name** (String, Optional) The name of the vault to retrieve.
- **uuid** (String, Optional) The UUID of the vault to retrieve.

### Read-only

- **description** (String, Read-only) The description of the vault.
- **id** (String, Read-only) The Terraform resource identifier for this item in the format `vaults/<vault_id>`


