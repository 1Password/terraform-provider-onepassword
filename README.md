<!-- Image sourced from https://blog.1password.com/introducing-secrets-automation/ -->
<img alt="" role="img" src="https://blog.1password.com/posts/2021/secrets-automation-launch/header.svg"/>

<div align="center">
  <h1>1Password Connect Terraform Provider</h1>
  <p>Use the 1Password Connect Terraform Provider to access and manage items in your 1Password vaults.</p>
  <a href="#-get-started">
    <img alt="Get started" src="https://user-images.githubusercontent.com/45081667/226940040-16d3684b-60f4-4d95-adb2-5757a8f1bc15.png" height="37"/>
  </a>
</div>

---

## ✨ Get started

- See the [examples](./examples/) directory for detailed examples.
- Detailed documentation for using this provider can be found on the [Terraform Registry docs](https://registry.terraform.io/providers/1Password/onepassword/latest/docs).

```tf
terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
      version = "~> 1.2.0"
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

## 🛠️ Contributing

For the contribution guidelines, see [CONTRIBUTING.md](/CONTRIBUTING.md).

Still not sure where or how to begin? We're happy to help! You can join the [Developer Slack workspace](https://join.slack.com/t/1password-devs/shared_invite/zt-1halo11ps-6o9pEv96xZ3LtX_VE0fJQA), and ask us any questions there.

## 💙 Community & Support

- File an [issue](https://github.com/1Password/terraform-provider-onepassword/issues) for bugs and feature requests.
- Join the [Developer Slack workspace](https://join.slack.com/t/1password-devs/shared_invite/zt-1halo11ps-6o9pEv96xZ3LtX_VE0fJQA).
- Subscribe to the [Developer Newsletter](https://1password.com/dev-subscribe/).

## 🔐 Security

1Password requests you practice responsible disclosure if you discover a vulnerability.

Please file requests via [**BugCrowd**](https://bugcrowd.com/agilebits).

For information about security practices, please visit the [1Password Bug Bounty Program](https://bugcrowd.com/agilebits).
