data "onepassword_items" "example" {
  vault = "your-vault-id"
  # Each element may be an item title or an item UUID, and is used verbatim as the map key below.
  titles = ["DB_PASSWORD", "API_KEY", "abcdefghijklmnopqrstuvwxyz"]
}

output "db_password" {
  value     = data.onepassword_items.example.credentials["DB_PASSWORD"]
  sensitive = true
}

# An item requested by UUID is keyed by that UUID, not by its title.
output "item_by_uuid" {
  value     = data.onepassword_items.example.credentials["abcdefghijklmnopqrstuvwxyz"]
  sensitive = true
}

output "api_key_category" {
  value = data.onepassword_items.example.items["API_KEY"].category
}
