---
page_title: "onepassword_items Data Source - onepassword"
subcategory: ""
description: |-
  Use this to batch-read multiple items from a vault. Returns item details and a convenience credentials map.
---

# onepassword_items (Data Source)

Use this to batch-read multiple items from a vault by their titles or UUIDs. This is more efficient than using multiple `onepassword_item` data sources with `for_each`, as it reduces Terraform configuration boilerplate and provides a convenient `credentials` map for the common case of reading secret values.

## Example Usage

```terraform
data "onepassword_items" "secrets" {
  vault  = "your-vault-id"
  titles = ["DB_PASSWORD", "API_KEY", "JWT_SECRET"]
}

# Access individual credentials
output "db_password" {
  value     = data.onepassword_items.secrets.credentials["DB_PASSWORD"]
  sensitive = true
}

# Access full item details
output "api_key_category" {
  value = data.onepassword_items.secrets.items["API_KEY"].category
}
```

## Schema

### Required

- `vault` (String) The UUID of the vault the items are in.
- `titles` (List of String) A list of item titles (or UUIDs) to retrieve from the vault.

### Read-Only

- `credentials` (Map of String, Sensitive) A map from item title to its primary credential value. For API credential items, this is the `credential` field. For all other items, this is the `password` field.
- `id` (String) The Terraform resource identifier for this data source.
- `items` (Attributes Map) A map from item title to item details. (see [below for nested schema](#nestedatt--items))

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Read-Only:

- `category` (String) The category of the item.
- `credential` (String, Sensitive) (Only applies to the API credential category) API credential for this item.
- `id` (String) The UUID of the item.
- `note_value` (String, Sensitive) Secure Note value.
- `password` (String, Sensitive) Password for this item.
- `tags` (List of String) An array of strings of the tags assigned to the item.
- `title` (String) The title of the item.
- `url` (String) The primary URL for the item.
- `username` (String) Username for this item.
