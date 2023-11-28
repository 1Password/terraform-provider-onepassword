[//]: # (START/LATEST)
# Latest

## Features
  * A user-friendly description of a new feature. {issue-number}

## Fixes
 * A user-friendly description of a fix. {issue-number}

## Security
 * A user-friendly description of a security fix. {issue-number}

---

[//]: # (START/v1.3.0)
# v1.3.0

## Features
  * Service Accounts support. {#79}
  * Add debugging support. {#102}

## Fixes
 * Fix item creation with sections. {#96}

## Security
 * Bump google.golang.org/grpc from 1.56.2 to 1.56.3 {#104}

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

Initial 1Password Connect Terraform Provider release

## Features:

- Importing existing items from 1Password Vault
- Creating new items in a 1Password Vault
- Updating existing item resources

---
