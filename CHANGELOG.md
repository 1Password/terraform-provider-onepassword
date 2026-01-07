[//]: # (START/LATEST)
# Latest

## Features
  * A user-friendly description of a new feature. {issue-number}

## Fixes
 * A user-friendly description of a fix. {issue-number}

## Security
 * A user-friendly description of a security fix. {issue-number}

---

[//]: # (START/v3.0.2)
# v3.0.2

## Fixes
 * Docs are updated to not mention specific provider version. {#297}
 * `purpose` property is removed from the section field. {#251}
 * Item data source correctly fetches the item with provided vault name when using service account or account to authenticate. {#306}

---

[//]: # (START/v3.0.1)
# v3.0.1

## Fixes

  * Provider produces consistent results after apply. {#223, #170}
  * Database item doesn't throw an error anymore. {#215}
  * Provider doesn't throw an error for sensitive attibutes. {#185}
  * SSH private keys in OpenSSH format are properly handled. {#286}
  * Provider reads API credential items correctly. {#287}
  * Provider properly handles string values in sections. {#214}

---

[//]: # (START/v3.0.0)
# v3.0.0

## 🔴 Breaking Changes
  * Remove `letters` option from password recipes. Letters are now always included in generated passwords and cannot be disabled. Configurations using `letters` in `password_recipe` will result in an error. {#256}
  * The `account` field should now be set to the account name. Account name appears at the top of the left sidebar in the 1Password desktop app.
  * Users who use biometric authentication should configure 1Password desktop app. {#270}

## Features
  * Enable provider to run using Terraform Stacks on HCP Terraform with a self-hosted agent. {#227}
  * Enable provider to run on Terraform Cloud. {#141}
  * `connect_url` and `connect_token` configuration parameters are available now. These are more specific alternatives to `url` and `token` for Connect authentication. The original `url` and `token` parameters continue to be supported but are set for deprecation. {#265}

## Fixes
  * `1password cli` is not required anymore to use the provider. {#148, #206}
  * Biometric unlock doesn't pop up multiple times anymore. {#140, #130}
  * Provider re-creates item when it's missing in 1Password vault. {#97}
  * No 504 Gateway Timeout errors anymore for newly created items. {#5}
  * Fix field reference from `label` to `id` for item data source. {#213}
  * `password_recipe` properly generates passwords now. {#129}
  * Testing documentation now includes make commands and setup instructions about how to run e2e tests. {#269}
  * README and documentation now to references to 1Password developer portal for more info. {#266}

## Security
  * Update `golang.org/x/crypto` from 0.39.0 to 0.45.0 to address security vulnerabilities. {#276}



---

[//]: # (START/v2.2.1)
# v2.2.1

## Fixes
 * Add testing documentation. {#242}
 * Eventual consistency for Connect client. {#246}
 * Fix eventual consultancy issue in Connect's item get implementation. {#244}
 * Fix item creation with incorrect date when using Connect. {#247}
 * Trimming trailing newline from `op read` command output. {#245}

---

[//]: # (START/v2.2.0)
# v2.2.0

## Features
  * Add `private_key_openssh` property to Item Data Source that returns SSH private key in OpenSSH format. {#189}

## Security
 * Address dependabot alerts and update Go version. {#226}

---

[//]: # (START/v2.1.2)
# v2.1.2

## Fixes
 * Export provider initialization function. {#196}

---

[//]: # (START/v2.1.1)
# v2.1.1

## Fixes
 * Update Go mod name. {#193}

## Security
 * Update dependencies with security vulnerabilities to patched versions. {#192}

---

[//]: # (START/v2.1.0)
# v2.1.0

## Features
  * Add support for Document Item category in item data source. {#171}
  * Add support for getting file attachments of an item in item data source. {#171}
  * Add support for getting an API Credential item's credential value in item data source. {#151}
  * Add support for SSH Key Item category in item data source. {#158}

## Fixes
 * Set password to null if not set. {#173}
 * Throw a better error message when item creation fails. {#174}
 * Improve examples and documentation. {#174}

---

[//]: # (START/v2.0.0)
# v2.0.0

## Features
 * Added support for `Secure Note` items. {#149}
 * Added `note_value` attribute representing a 1Password Item's `notes` field. {#57}

## Fixes
 * The data handling is more robust, making it less prone to errors and inconsistencies. {#157,#146}
 * CLI and Connect clients now have a more consistent behavior.
 * Fields of type `OTP` are better handled when user provides a custom ID for them. Terraform will throw an error if the custom ID doesn't have the `TOTP_` prefix, which is required for this field type.
 * The values that are generated will only show in the plan to be recomputed when the recipe is changed or the value is explicitly set.
 * When fetching Database items from 1Password, the `server` field (previously known as `hostname`) will populate the Terraform `hostname` attribute. This ensures that the data from new Database items is mapped as expected. {#76}
 * Vault description is now fetched when getting a vault from 1Password by name and the provider was configured to use the CLI client.
 * Generated values (using a recipe) are now regenerated when the recipe is changed.
 * Tag ordering mismatch between Terraform state and 1Password no longer causes a change if the tags are the same. The mismatch can be caused by 1Password storing the tags in alphabetical order. {#155}

## Security
 * Migration to Terraform Provider Framework addressed an issue in the terraform-plugin-sdk where it is possible that sensitive data pulled from 1Password items can be shown in plaintext when a user runs `terraform plan`. This only affects the sensitive data pulled from custom sections within 1Password items that aren’t marked as sensitive in the terraform plan. This also applies to third-party providers that don’t treat the data as sensitive. {#167}

---

[//]: # (START/v1.4.3)
# v1.4.3

## Fixes
 * Pass proper user agent info to the CLI. {#124}

---

[//]: # (START/v1.4.2)
# v1.4.2

## Fixes
 * Field of type 'DATE' updates item even if there were no changes. {#137}

## Security
 * Update dependencies with security vulnerabilities to patched versions. {#144}

---

[//]: # (START/v1.4.1)
# v1.4.1

## Features
 * Using provider on Terraform Cloud. {#116}

## Fixes
 * Terraform cannot create items with the password we provide in the code. {#128}

---

[//]: # (START/v1.4.1-beta01)
# v1.4.1-beta01

## Fixes
* Using provider on Terraform Cloud. {#116}

---

[//]: # (START/v1.4.0)
# v1.4.0

## Features
  * Authenticate 1Password CLI with biometric unlock using user account. {#113}

## Fixes
 * Retry CLI request in case of 409 Conflict error. {#108}
 * Update documentation. {#115}

---

[//]: # (START/v1.3.1)
# v1.3.1

## Fixes
 * Update documentation to mention that the provider supports Service Accounts. {#106}

---

[//]: # (START/v1.3.0)
# v1.3.0

## Features
  * Add Service Accounts support. Credits to @tim-oster for the contribution! {#79}
  * Add debugging support. {#102}

## Security
 * Update dependencies with security vulnerabilities to patched versions. {#104, #112}

---

[//]: # (START/v1.2.1)
# v1.2.1

## Fixes
 * Fix item creation with sections. {#96}

---

[//]: # (START/v1.2.0)
# v1.2.0

## Features
  * Updating go version to 1.20
  * Updating to use version 1.5.1 of the Connect SDK.

## Fixes
 * Improved sanitization for use with Github action.
 * Terraform provider no longer lowercases item label. {#59}

## Security
 * Updated dependencies with secuirty vulnerbilities to patched versions

---

[//]: # (START/v1.1.4)
# v1.1.4

## Fixes
 * Fix (T)OTP field type. {#54}

---

[//]: # (START/v1.1.3)
# v1.1.3

## Fixes
 * Setting the provider's `token` field through Terraform's built-in prompt no longer leads to an error about the `url` not beign set. {#46}
 * The purpose of the `id` and `uuid` fields of the item and vault data-source is now correctly described in the docs. {#42}
 * The `tags` field for the item data-source is now correctly identified as an output.

---

[//]: # (START/v1.1.2)
# v1.1.2
## Fixes
 * Improve error messages.

[//]: # (START/v1.1.1)
# v1.1.1

## Features
 * This release includes a binary for `darwin/arm64`. {#35}

---

[//]: # (START/v1.1.0)
# v1.1.0

## Features
 * Adds the `onepassword_vault` data source that can be used to look up a vault by its name or uuid. {#25}
 * The `onepassword_item` data source can now be used by setting the `title` instead of the `uuid` field. {#25}
 * The documentation now clearly mentions that the Connect Token can also be provided thorugh `$OP_CONNECT_TOKEN`.

## Fixes
 * The `id` and `uuid` fields of the `onepassword_item` resource are now correctly designated as outputs.

---

[//]: # (START/v1.0.2)
# v1.0.2

## Features
 * Documentation for the provider is now published on the Terraform Registry. {#8}

---

[//]: # "START/v1.0.1"

# v1.0.1

## Fixes

- Tags set in the `tags` field now correctly get set on items in your vault. {#15}
- Changing the category of an item no longer results in the contents of your item no longer being visible. {#13}
- Changing the vault of an item no longer leads to an error that the item cannot be found.

---

[//]: # "START/v1.0.0"

# v1.0.0

## Features

- Update to the lastest Connect SDK
- Includes a Terraform User Agent in all Connect API requests

---

[//]: # "START/v0.2.0"

# v0.2.0

Support custom sections and fields for Login, Password, and Database Items

## Features:

- Add support for defining Sections and Fields for Items
- Access sections and fields through `onepassword_item` data source

---

[//]: # "START/v0.1.0"

# v0.1.0

Initial 1Password Terraform provider release

## Features:

- Importing existing items from 1Password Vault
- Creating new items in a 1Password Vault
- Updating existing item resources

---
