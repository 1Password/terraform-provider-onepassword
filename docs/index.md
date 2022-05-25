---
layout: ""
page_title: "Provider: 1Password"
description: |-
  Use the 1Password Connect Terraform Provider to reference, create, or update logins, password and database items in your 1Password Vaults.
---

# 1Password Connect Terraform Provider

Use the 1Password Connect Terraform Provider to reference, create, or update items in your existing vaults using [1Password Secrets Automation](https://1password.com/secrets).

## Example Usage

Connecting to 1Password Connect

```terraform
provider "onepassword" {
  url = "http://localhost:8080"
}
```

Connecting to 1Password local CLI

```terraform
provider "onepassword" {
  account  = "username"
  password = "password"
}
```

## Schema

### Optional

- **token** (String, Optional) A valid token for your 1Password Connect API. Can also be sourced from OP_CONNECT_TOKEN.

### Optional

- **url** (String, Optional) The HTTP(S) URL where your 1Password Connect API can be found. Must be provided through the OP_CONNECT_HOST environment variable if this attribute is not set.

### Optional

- **account** (String, Optional) Account to use for the 1Password CLI. Can also be sourced from OP_ACCOUNT

### Optional

- **password** (String, Optional) Password to use for the 1Password CLI. Can also be sourced from OP_PASSWORD