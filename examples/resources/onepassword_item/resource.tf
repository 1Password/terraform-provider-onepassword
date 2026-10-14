# Example using section_map (map-based sections for direct field access)
resource "onepassword_item" "example_with_map" {
  vault = "your-vault-id"

  title    = "Example Item with Section Map"
  category = "login"

  section_map = {
    "credentials" = {
      field_map = {
        "api_key" = {
          type  = "CONCEALED"
          value = "my-secret-api-key"
        }
        "api_endpoint" = {
          type  = "URL"
          value = "https://api.example.com"
        }
      }
    }
    "metadata" = {
      field_map = {
        "environment" = {
          type  = "STRING"
          value = "production"
        }
        "created_by" = {
          type  = "STRING"
          value = "terraform"
        }
      }
    }
  }
}

# Example using section (list-based sections)
resource "onepassword_item" "example_with_list" {
  vault = "your-vault-id"

  title    = "Example Item with Section List"
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

# Example generating an SSH key pair (ssh_key category)
resource "onepassword_item" "example_ssh_key" {
  vault = "your-vault-id"

  title        = "Deploy Key"
  category     = "ssh_key"
  ssh_key_type = "ed25519" # or "rsa"
  # ssh_key_bits = 4096    # only used for rsa (2048-4096, default 2048)
}

# Example API credential item
resource "onepassword_item" "example_api_credential" {
  vault = "your-vault-id"

  title      = "Example API Credential"
  category   = "api_credential"
  username   = "service-account"
  credential = "op://vault/item/credential"
  valid_from = "2026-01-01"
  filename   = "credentials.json"
}

# Example server item with admin console credentials
resource "onepassword_item" "example_server" {
  vault = "your-vault-id"

  title                  = "Example Server"
  category               = "server"
  username               = "deploy"
  admin_console_url      = "https://console.example.com"
  admin_console_username = "admin"
  admin_console_password = "op://vault/item/password"
}

# Example memorable and PIN password recipes
resource "onepassword_item" "example_memorable" {
  vault = "your-vault-id"

  title    = "Example Memorable Password"
  category = "password"

  password_recipe {
    type       = "memorable"
    word_count = 3
    separator  = "hyphens"
  }
}

resource "onepassword_item" "example_pin" {
  vault = "your-vault-id"

  title    = "Example PIN"
  category = "password"

  password_recipe {
    type   = "pin"
    length = 6
  }
}
