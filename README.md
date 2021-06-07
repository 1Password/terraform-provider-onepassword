# 1Password Connect Terraform Provider

Use the 1Password Connect Terraform Provider to reference, create, or update items in your 1Password Vaults.

## Usage

Detailed documentation for using this provider can be found [here](https://registry.terraform.io/providers/1Password/onepassword/latest/docs).

```tf
terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
      version = "~> 1.0.0"
    }
  }
}

provider "onepassword" {
  url = "http://localhost:8080"
}

variable "vault_id" {}

resource "onepassword_item" "demo_login" {
  vault = var.vault_id

  title    = "Demo Terraform Login"
  category = "password"

  username = "demo-username"

  password_recipe {
    length  = 40
    symbols = false
  }
}
```

See the [examples](./examples/) directory for a full example.

## Contributing

Detailed documentation for contributing to the 1Password Terraform provider can be found [here](./CONTRIBUTING.md).
