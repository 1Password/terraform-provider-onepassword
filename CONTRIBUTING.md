# Contributing

Thanks for your interest in contributing to the 1Password Connect Terraform Provider project! 🙌 We appreciate your time and effort. Here are some guidelines to help you get started.

## Building

Run the following command to build the 1Password Connect Terraform Provider:

```sh
go build .
```

This will create the `terraform-provider-onepassword` binary.

## Testing the Provider

To run the Go tests and check test coverage run the following command:

```sh
go test -v ./... -cover
```

## Installing plugin locally

Refer to the following sections of the Terraform's "Custom Framework Providers" tutorial to install this plugin locally:

- [Prepare Terraform for local provider install](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
- [Locally install provider and verify with Terraform](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#locally-install-provider-and-verify-with-terraform)

## Using plugin locally

In your Terraform configuration you will need to specify the `op` plugin with:

```tf
terraform {
  required_providers {
    onepassword = {
      source  = "1Password/onepassword"
      version = "~> 1.2.0"
    }
  }
}

provider "onepassword" {
  url = "http://<1Password Connect API Hostname>"
}
```

After copying a newly-built version of the provider to the plugins directory you will have to run `terraform init` again. If you forget to do this then Terraform will error out and tell you to do so.

## Generating Documentation

Documentation is generated for the provider using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs). This plugin uses the schema `Description` field in conjunction with the contents of the `/templates` and `/examples` folders to generate the `/docs` content.

To regenerate the `/docs` Markdown run:

```sh
go generate
```
