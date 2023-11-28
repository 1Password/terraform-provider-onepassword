# Contributing

Thanks for your interest in contributing to the 1Password Connect Terraform Provider project! 🙌 We appreciate your time and effort. Here are some guidelines to help you get started.

## Building

Run the following command to build the 1Password Connect Terraform Provider:

```sh
$ go build .
```

This will create the `terraform-provider-onepassword` binary.

## Testing the Provider

To run the Go tests and check test coverage run the following command:

```sh
$ go test -v ./... -cover
```

## Installing the Provider Locally

Refer to the following sections of the Terraform's "Custom Framework Providers" tutorial to install this provider locally:

- [Prepare Terraform for local provider install](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
- [Locally install provider and verify with Terraform](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#locally-install-provider-and-verify-with-terraform)

## Using the Provider Locally

In your Terraform configuration you will need to specify the `onepassword` provider with:

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

## Debugging

Ensure that the `dev_overrides` block was added to your `~/.terraformrc` file, using `"1Password/onepassword"` as the source. Refer to the [Installing the Plugin Locally](#installing-the-plugin-locally) section for instructions.

Build the provider without optimizations enabled:

```sh
$ go build -gcflags="all=-N -l" .
```

Start a Delve debugging session:

```sh
$ dlv debug . -- --debug
Type 'help' for list of commands.
(dlv) continue
```

**Note**: Editors like GoLand can be configured to start a debugging session as well. Just be sure to pass the `--debug` flag as a program argument.

If a debugging session was started properly, the provider should print the following output to `stdout`: 

```sh
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

    TF_REATTACH_PROVIDERS='{"1Password/onepassword":{"Protocol":"grpc","Pid":3382870,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin713096927"}}}'

```

Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output. Either export it, or prefix every Terraform command with it, and run Terraform as usual. Any breakpoints you have set will halt execution and show you the current variable values.

## Generating Documentation

Documentation is generated for the provider using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs). This plugin uses the schema `Description` field in conjunction with the contents of the `/templates` and `/examples` folders to generate the `/docs` content.

To regenerate the `/docs` Markdown run:

```sh
$ go generate
```

## Sign Your Commits

To get your PR merged, we require you to sign your commits. Fortunately, this has become very easy to [set up](https://developer.1password.com/docs/ssh/git-commit-signing/)!
