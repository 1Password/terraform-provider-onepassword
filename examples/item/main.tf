terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
    }
  }
}

data "onepassword_vault" "demo_vault" {
  name = var.demo_vault
}

resource "onepassword_item" "demo_login" {
  vault = data.onepassword_vault.demo_vault.uuid

  title    = "Demo Terraform Login Item"
  category = "login"
  username = "test@example.com"

  tags = ["Terraform", "Automation"]

  password_recipe {
    length  = 32
    digits  = false
    symbols = false
  }

  note_value = "An item created with the 1Password Terraform provider"
}

resource "onepassword_item" "demo_password" {
  vault = data.onepassword_vault.demo_vault.uuid

  title    = "Demo Terraform Password Item"
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

resource "onepassword_item" "demo_db" {
  vault    = data.onepassword_vault.demo_vault.uuid
  category = "database"
  type     = "mysql"

  title    = "Demo Terraform Database Item"
  username = "root"

  database = "Example MySQL Instance"
  hostname = "localhost"
  port     = 3306
}

resource "onepassword_item" "demo_secure_note" {
  vault = data.onepassword_vault.demo_vault.uuid

  title    = "Demo Terraform Secure Note Item"
  category = "secure_note"

  note_value = <<EOT
  Welcome to the Terraform world! 🤩
  This was an item created with the 1Password Terraform provider.
  EOT
}

resource "onepassword_item" "demo_sections" {
  vault = data.onepassword_vault.demo_vault.uuid

  title    = "Demo Terraform Item with Sections"
  category = "login"
  username = "test@example.com"

  section {
    label = "Terraform Section"

    field {
      label = "API_KEY"
      type  = "CONCEALED"
      value = "2Federate2!"
    }

    field {
      label = "HOSTNAME"
      value = "example.com"
    }

    field {
      label = "API key Expiry"
      type  = "DATE"
      value = "2024-01-31"
    }
  }

  section {
    label = "Terraform Second Section"

    field {
      label = "App Specific Password"
      type  = "CONCEALED"

      password_recipe {
        length  = 40
        symbols = false
      }
    }

    field {
      label = "User"
      value = "demo"
    }
  }
}

# Example of a Data Source Item with multiple sections and fields.
# Uncomment it once the item above has been created to see an example of a Data Source
# data "onepassword_item" "example" {
#   vault = data.onepassword_vault.demo_vault.uuid
#   uuid  = onepassword_item.demo_sections.uuid
# }
