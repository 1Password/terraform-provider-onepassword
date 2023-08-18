# Contributing

Thanks for your interest in contributing to the 1Password Connect Terraform Provider project! 🙌 We appreciate your time and effort. Here are some guidelines to help you get started.

## Building

Run the following command to build the 1Password Connect Terraform Provider:

```sh
go build .
```

This will create the `terraform-provider-onepassword` binary.

## Testing the Provider

To run the Go tests and check test coverage run the following:

```sh
go test -v ./... -cover
```

## Generating Documentation

Documentation is generated for the provider using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs). This plugin uses the schema `Description` field in conjunction with the contents of the `/templates` and `/examples` folders to generate the `/docs` content.

To regenerate the `/docs` Markdown run:

```sh
go generate
```

## Installing plugin locally

To install the binary it must be copied to the appropriate Terraform plugin directory (note that you may need to create these directories first). You can refer to the [Terraform 0.13 plugin docs](https://www.hashicorp.com/blog/automatic-installation-of-third-party-providers-with-terraform-0-13) for more information.

The current version of the provider is considered to be 0.2 and `darwin_amd64` should match your machines operating system and architecture in the format `$OS_$ARCH`. For example, macOS is `darwin_amd64` and Linux is `linux_amd64`.

```sh
mkdir -p ~/.terraform.d/plugins/github.com/1Password/onepassword/0.2/darwin_amd64/
cp ./terraform-provider-onepassword ~/.terraform.d/plugins/github.com/1Password/onepassword/0.2/darwin_amd64/terraform-provider-onepassword
```

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
