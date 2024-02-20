terraform {
  required_providers {
    onepassword = {
      source  = "1Password/onepassword"
      version = "~> 1.3.0"
    }
  }
}

provider "onepassword" {
  url = "http://localhost:8080"
}

resource "onepassword_item" "demo_password" {
  vault = var.demo_vault

  title    = "Demo Password Recipe"
  category = "password"

  password_recipe {
    length  = 40
    symbols = false
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

resource "onepassword_item" "demo_sections" {
  vault = var.demo_vault

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

resource "onepassword_item" "demo_date" {
  vault = var.demo_vault
  title = "Important Date"

  section {
    label = "Date"
    field {
      label = "date"
      type  = "DATE"
      value = "2024-01-31"
    }
  }
}

# Example of a Data Source Item with multiple sections and fields.
# Uncomment it once the item above has been created to see an example of a Data Source
# data "onepassword_item" "example" {
#   vault = var.demo_vault
#   uuid  = onepassword_item.demo_sections.uuid
# }
