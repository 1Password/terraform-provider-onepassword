# Creating Login, Password, and Database 1Password Items

For more details check out [1Password Terraform Provider documentation](https://developer.1password.com/docs/terraform/).

## Create the Items

From the `examples/directory` run:

```sh
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding 1password/onepassword versions matching "~> 1.0.0"...
- Installing 1password/onepassword v1.0.0...
- Installed 1password/onepassword v1.0.0 (signed by a HashiCorp partner, key ID 6681876AE08DC4BF)

Partner and community providers are signed by their developers.
If you'd like to know more about provider signing, you can read about it here:
https://www.terraform.io/docs/cli/plugins/signing.html

Terraform has created a lock file .terraform.lock.hcl to record the provider
selections it made above. Include this file in your version control repository
so that Terraform can guarantee to make the same selections by default when
you run "terraform init" in the future.

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.

$ terraform apply


An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # onepassword_item.demo_db will be created
  + resource "onepassword_item" "demo_db" {
      + category = "database"
      + database = "Example MySQL Instance"
      + hostname = "localhost"
      + id       = (known after apply)
      + password = (sensitive value)
      + port     = "3306"
      + title    = "Demo TF Database"
      + type     = "mysql"
      + username = "root"
      + uuid     = (known after apply)
      + vault    = "<TF_VAR_demo_vault>"
    }

  # onepassword_item.demo_login will be created
  + resource "onepassword_item" "demo_login" {
      + category = "login"
      + id       = (known after apply)
      + password = (sensitive value)
      + title    = "Demo Terraform Login"
      + username = "test@example.com"
      + uuid     = (known after apply)
      + vault    = "<TF_VAR_demo_vault>"
    }

  # onepassword_item.demo_password will be created
  + resource "onepassword_item" "demo_password" {
      + category = "password"
      + id       = (known after apply)
      + password = (sensitive value)
      + title    = "Demo Password Recipe"
      + uuid     = (known after apply)
      + vault    = "<TF_VAR_demo_vault>"

      + password_recipe {
          + digits  = true
          + length  = 40
          + symbols = false
        }
    }

Plan: 3 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

onepassword_item.demo_db: Creating...
onepassword_item.demo_login: Creating...
onepassword_item.demo_password: Creating...
onepassword_item.demo_password: Creation complete after 0s [id=vaults/<TF_VAR_demo_vault>/items/<New Item UUID>]
onepassword_item.demo_db: Creation complete after 1s [id=vaults/<TF_VAR_demo_vault>/items/<New Item UUID>]
onepassword_item.demo_login: Creation complete after 1s [id=vaults/<TF_VAR_demo_vault>/items/<New Item UUID>]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
```

## Destroy the Items

Clean up all the resources that were created with:

```sh
$ terraform destroy

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # onepassword_item.demo_db will be destroyed
  - resource "onepassword_item" "demo_db" {
      - category = "database" -> null
      - database = "Example MySQL Instance" -> null
      - hostname = "localhost" -> null
      - id       = "vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>" -> null
      - password = (sensitive value)
      - port     = "3306" -> null
      - title    = "Demo TF Database" -> null
      - type     = "mysql" -> null
      - username = "root" -> null
      - uuid     = "<Item UUID from Create>" -> null
      - vault    = "<TF_VAR_demo_vault>" -> null
    }

  # onepassword_item.demo_login will be destroyed
  - resource "onepassword_item" "demo_login" {
      - category = "login" -> null
      - id       = "vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>" -> null
      - password = (sensitive value)
      - title    = "Demo Terraform Login" -> null
      - username = "test@example.com" -> null
      - uuid     = "<Item UUID from Create>" -> null
      - vault    = "<TF_VAR_demo_vault>" -> null
    }

  # onepassword_item.demo_password will be destroyed
  - resource "onepassword_item" "demo_password" {
      - category = "password" -> null
      - id       = "vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>" -> null
      - password = (sensitive value)
      - title    = "Demo Password Recipe" -> null
      - uuid     = "<Item UUID from Create>" -> null
      - vault    = "<TF_VAR_demo_vault>" -> null

      - password_recipe {
          - digits  = true -> null
          - length  = 40 -> null
          - symbols = false -> null
        }
    }

Plan: 0 to add, 0 to change, 3 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

onepassword_item.demo_login: Destroying... [id=vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>]
onepassword_item.demo_db: Destroying... [id=vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>]
onepassword_item.demo_password: Destroying... [id=vaults/<TF_VAR_demo_vault>/items/<Item UUID from Create>]
onepassword_item.demo_login: Destruction complete after 0s
onepassword_item.demo_db: Destruction complete after 0s
onepassword_item.demo_password: Destruction complete after 0s
```
