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

# Example storing a file in a document item.
# Requires service account or desktop app authentication (not supported with 1Password Connect).
resource "onepassword_item" "example_document" {
  vault = "your-vault-id"

  title    = "Example TLS Certificate"
  category = "document"

  document = {
    name    = "server.pem"
    content = file("${path.module}/server.pem")
    # For binary files, use content_base64 instead:
    # content_base64 = filebase64("${path.module}/keystore.p12")
  }
}
