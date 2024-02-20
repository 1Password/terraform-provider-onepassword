resource "onepassword_item" "demo_password" {
  vault = var.demo_vault

  title    = "Demo Password Recipe"
  category = "password"

  password_recipe {
    length  = 40
    symbols = false
  }

  section {
    label = "Credential metadata"

    field {
      label = "Expiration"
      type  = "DATE"
      value = "2024-01-31"
    }
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
