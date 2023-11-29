# Contributing

Thanks for your interest in contributing to the 1Password Terraform provider project! 🙌 We appreciate your time and effort. Here are some guidelines to help you get started.

## Building

Run the following command to build the 1Password Terraform provider:

```sh
go build .
```

This will create the `terraform-provider-onepassword` binary.

## Testing the Provider

To run the Go tests and check test coverage run the following command:

```sh
go test -v ./... -cover
```

## Installing the provider locally

Refer to the following sections of the Terraform's "Custom Framework Providers" tutorial to install this provider locally:

- [Prepare Terraform for local provider install](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
- [Locally install provider and verify with Terraform](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#locally-install-provider-and-verify-with-terraform)

## Using the provider locally

You must specify the `onepassword` provider in your Terraform configuration:

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

After copying a newly built version of the provider to the plugins directory, you need to run `terraform init` again. Otherwise, Terraform returns an error.

## Debugging

Make sure you add the `dev_overrides` block to your `~/.terraformrc` file (using `"1Password/onepassword"` as the source). For instructions, refer to the [Installing the provider locally](#installing-the-provider-locally).

Build the provider without optimizations enabled:

```sh
go build -gcflags="all=-N -l" .
```

Start a Delve debugging session:

```sh
dlv debug . -- --debug
Type 'help' for list of commands.
(dlv) continue
```

**Note**: You can also configure editors like GoLand to start a debugging session by passing the `--debug` flag as a program argument.

If a debugging session was starts correctly, the provider prints the following output to `stdout`:

```sh
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

    TF_REATTACH_PROVIDERS='{"1Password/onepassword":{"Protocol":"grpc","Pid":3382870,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin713096927"}}}'

```

Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output. Either export it, or prefix every Terraform command with it, and run Terraform as usual. Any breakpoints you have set will halt execution and show you the current variable values.

## Generating documentation

Documentation is generated for the provider using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs). This plugin uses the schema `Description` field in conjunction with the contents of the `/templates` and `/examples` folders to generate the `/docs` content.

To regenerate the `/docs` Markdown run:

```sh
go generate
```

## Sign your commits

To get your PR merged, we require you to sign your commits.

### Sign commits with `1Password`

You can also sign commits using 1Password, which lets you sign commits with biometrics without the signing key leaving the local 1Password process.

Learn how to use [1Password to sign your commits](https://developer.1password.com/docs/ssh/git-commit-signing/).


### Sign commits with `ssh-agent`

Follow the steps below to set up commit signing with `ssh-agent`:

1. Generate an SSH key and add it to ssh-agent
2. Add the SSH key to your GitHub account
3. Configure git to use your SSH key for commit signing

### Sign commits `gpg`

Follow the steps below to set up commit signing with `gpg`:

1. Generate a GPG key
2. Add the GPG key to your GitHub account
3. Configure git to use your GPG key for commit signing