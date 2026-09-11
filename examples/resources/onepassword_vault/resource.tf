resource "onepassword_vault" "example" {
  name                = "Terraform managed vault"
  description         = "Managed by Terraform"
  allow_admins_access = true

  group_permissions {
    group_id = "2hinhvewsxob5p3he2cr2ylihy"
    permissions = [
      "read_items",
      "reveal_item_password",
      "create_items",
      "update_items",
    ]
  }

  group_permissions {
    group_id = "3gjoowf6ofrcjldq5z7v6h3cde"
    permissions = [
      "read_items",
    ]
  }
}
