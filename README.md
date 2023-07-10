# 1Password Connect Terraform Provider

Use the 1Password Connect Terraform Provider to reference, create, or update items in your 1Password Vaults.

## Usage

Detailed documentation for using this provider can be found on the [Terraform Registry docs](https://registry.terraform.io/providers/1Password/onepassword/latest/docs).

```tf
terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
      version = "~> 1.1.2"
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

## 🛠️ Contributing

For the contribution guidelines, see [CONTRIBUTING.md](/CONTRIBUTING.md).

Still not sure where or how to begin? We're happy to help! You can:

- Join the [Developer Slack workspace](https://join.slack.com/t/1password-devs/shared_invite/zt-1halo11ps-6o9pEv96xZ3LtX_VE0fJQA), and ask us any questions there.
