resource "onepassword_item" "example" {
  vault = "your-vault-id"

  title    = "Example Item Title"
  category = "login"

  password_recipe {
    length  = 40
    symbols = false
  }

  section {
    label = "Example section"

    field {
      label = "Example field"
      type  = "DATE"
      value = "2024-01-31"
    }
  }
}
