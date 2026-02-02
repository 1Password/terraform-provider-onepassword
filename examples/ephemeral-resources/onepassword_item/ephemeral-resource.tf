# Example using ephemeral resource to retrieve item values
ephemeral "onepassword_item" "example" {
  vault = "your-vault-id"
  title = "your-item-title"
}

# Example using UUID instead of title
ephemeral "onepassword_item" "example_by_uuid" {
  vault = "your-vault-id"
  uuid  = "your-item-uuid"
}
