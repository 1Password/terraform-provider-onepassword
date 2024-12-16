resource "onepassword_item_share" "example" {
  vault = "your-vault-id"
  item  = "your-item-id"

  emails     = "person1@example.com,person2@example.com"
  expires_in = "3d"
}
