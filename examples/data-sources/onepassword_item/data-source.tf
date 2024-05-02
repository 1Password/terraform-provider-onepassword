data "onepassword_item" "example" {
  vault = data.onepassword_vault.example.uuid
  uuid  = onepassword_item.demo_sections.uuid
}
