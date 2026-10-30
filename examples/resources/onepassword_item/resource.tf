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

# Example writing a section field's value without storing it in state.
# Increment value_wo_version to write a new value; otherwise the value in 1Password is kept.
ephemeral "random_password" "signing_key" {
  length = 48
}

resource "onepassword_item" "example_with_write_only_field" {
  vault = "your-vault-id"

  title    = "Example Item with a Write-Only Field"
  category = "secure_note"

  section_map = {
    "credentials" = {
      field_map = {
        "signing_key" = {
          type             = "CONCEALED"
          value_wo         = ephemeral.random_password.signing_key.result
          value_wo_version = 1
        }
        "key_id" = {
          type  = "STRING"
          value = "key-01"
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
