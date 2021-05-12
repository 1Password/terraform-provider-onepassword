data "onepassword_item" "example" {
  vault = var.demo_vault
  uuid  = onepassword_item.demo_sections.uuid
}
