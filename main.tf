terraform {
  required_version = ">= 1.0"
  required_providers {
    onepassword = {
      source  = "1Password/onepassword"
    }
  }
}


data "onepassword_vault" "terraform" {
  name = "terraform-provider-acceptance-tests"
}


 resource "onepassword_item" "test_section_field_wo_initial" {
  vault    = data.onepassword_vault.terraform.uuid
  title    = "Test Section Field Write-Only - Initial"
  category = "login"
  username = "testuser@example.com"

  section {
    label = "Credentials Section"

    field {
      label            = "API Key"
      type             = "CONCEALED"
      value_wo         = "initial-api-key-1234566666666677"
      value_wo_version = 3
    }
  }
}
