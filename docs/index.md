---
layout: ""
page_title: "Provider: 1Password"
description: |-
  Use the 1Password Connect Terraform Provider to reference, create, or update logins, password and database items in your 1Password Vaults.
---

# 1Password Connect Terraform Provider

Use the 1Password Connect Terraform Provider to reference, create, or update items in your existing vaults using [1Password Secrets Automation](https://1password.com/secrets).

## Example Usage

```terraform
provider "onepassword" {
  url = "http://localhost:8080"
}
```

## Schema

### Required

- **token** (String, Required) A valid token for your 1Password Connect API. Can also be sourced from OP_CONNECT_TOKEN.

### Optional

- **url** (String, Optional) The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the the OP_CONNECT_HOST environment variable if this attribute is not set.
