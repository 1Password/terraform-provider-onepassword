package provider

import (
	"strings"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

const (
	terraformItemIDDescription = "The Terraform resource identifier for this item in the format `vaults/<vault_id>/items/<item_id>`."

	itemDataSourceDescription = "Use this to get details of an item by its vault uuid and either the title or the uuid of the item."
	itemEphemeralDescription  = "Use this to retrieve item values without storing them in Terraform state. Useful for providing sensitive values to write-only arguments or other ephemeral contexts."

	itemLookupUUIDDescription  = "The UUID of the item to retrieve. This field will be populated with the UUID of the item if the item it looked up by its title."
	itemLookupTitleDescription = "The title of the item to retrieve. This field will be populated with the title of the item if the item it looked up by its UUID."

	itemUUIDDescription                  = "The UUID of the item. Item identifiers are unique within a specific vault."
	vaultUUIDDescription                 = "The UUID of the vault the item is in."
	categoryDescription                  = "The category of the item."
	itemTitleDescription                 = "The title of the item."
	urlDescription                       = "The primary URL for the item."
	tagsDescription                      = "An array of strings of the tags assigned to the item."
	usernameDescription                  = "Username for this item."
	passwordDescription                  = "Password for this item."
	passwordWriteOnceDescription         = "A write-only password for this item. This value is not stored in the state and is intended for use with ephemeral values. **Note**: Write-only arguments require Terraform 1.11 or later."
	passwordWriteOnceVersionDescription  = "An integer that must be incremented to trigger an update to the 'password_wo' field."
	noteValueWriteOnceDescription        = "A write-only secure note value for this item. This value is not stored in the state and is intended for use with ephemeral values. **Note**: Write-only arguments require Terraform 1.11 or later."
	noteValueWriteOnceVersionDescription = "An integer that must be incremented to trigger an update to the 'note_value_wo' field."
	noteValueDescription                 = "Secure Note value."
	publicKeyDescription                 = "SSH Public Key for this item."
	privateKeyDescription                = "SSH Private Key in PKCS#8 for this item."
	privateKeyOpenSSHDescription         = "SSH Private key in OpenSSH format."
	credentialDescription                = "(Only applies to the API credential category) API credential for this item."
	validFromDescription                 = "(Only applies to the API credential category) The timestamp from which the API credential is valid."
	filenameDescription                  = "(Only applies to the API credential category) The filename associated with the API credential."

	dbHostnameDescription = "(Only applies to the database category) The address where the database can be found"
	dbDatabaseDescription = "(Only applies to the database category) The name of the database."
	dbPortDescription     = "(Only applies to the database category) The port the database is listening on."
	dbTypeDescription     = "(Only applies to the database category) The type of database."
	typeDescription       = "(Only applies to database and API credential categories) The type of database or API Credential."

	sshKeyTypeDescription           = "(Only applies to the SSH key category) The type of SSH key to generate. **Note**: 1Password Connect has limited support for SSH key items (see [1Password/connect#107](https://github.com/1Password/connect/issues/107)); the service account and desktop app integrations are recommended."
	sshKeyBitsDescription           = "(Only applies to the SSH key category) The size in bits of the RSA key to generate. Only used when `ssh_key_type` is `rsa`."
	fingerprintDescription          = "(Only applies to the SSH key category) The SHA256 fingerprint of the generated public key."
	sshKeyTypeOfDescription         = "(Only applies to the SSH key category) The type of the generated key, e.g. `Ed25519` or `RSA, 2048-bit`."
	adminConsoleURLDescription      = "(Only applies to the server category) The URL of the server's admin console."
	adminConsoleUsernameDescription = "(Only applies to the server category) The username for the server's admin console."
	adminConsolePasswordDescription = "(Only applies to the server category) The password for the server's admin console."
	networkNameDescription          = "(Only applies to the wireless router category) The name of the wireless network."
	serverAddressDescription        = "(Only applies to the wireless router category) The server / IP address of the router."
	wirelessSecurityDescription     = "(Only applies to the wireless router category) The wireless security type, e.g. `WPA2` or `WPA3`."
	wirelessPasswordDescription     = "(Only applies to the wireless router category) The password of the wireless network."
	licenseKeyDescription           = "(Only applies to the software license category) The license key."
	versionDescription              = "(Only applies to the software license category) The version of the licensed software."
	downloadLinkDescription         = "(Only applies to the software license category) The download page for the software."
	licensedToDescription           = "(Only applies to the software license category) The name of the license holder."
	registeredEmailDescription      = "(Only applies to the software license category) The email the license is registered to."

	sectionListDescription  = "A list of custom sections in an item. Cannot be used together with `section_map`. Use either `section` (list) or `section_map` (map), but not both."
	sectionMapDescription   = "A map of custom sections in an item, keyed by section label. This allows direct lookup of sections and their fields by label. Cannot be used together with `section`. Use either `section` (list) or `section_map` (map), but not both."
	sectionIDDescription    = "A unique identifier for the section."
	sectionLabelDescription = "The label for the section."

	documentFileListDescription  = "A list of files attached to the document item."
	fileListDescription          = "A list of files attached to the section."
	fileMapDescription           = "A map of files attached to the section, keyed by file label."
	fileIDDescription            = "The UUID of the file."
	fileNameDescription          = "The name of the file."
	fileContentDescription       = "The content of the file."
	fileContentBase64Description = "The content of the file in base64 encoding. (Use this for binary files.)"

	fieldListDescription  = "A list of custom fields in the section"
	fieldMapDescription   = "A map of custom fields in the section, keyed by field label."
	fieldIDDescription    = "A unique identifier for the field."
	fieldLabelDescription = "The label for the field."
	fieldTypeDescription  = "The type of value stored in the field."
	fieldValueDescription = "The value of the field."

	passwordRecipeDescription       = "The recipe used to generate a new value for a password."
	passwordTypeDescription         = "The type of password to generate."
	passwordLengthDescription       = "The length of the password to be generated. Applies to `random` and `pin` types."
	passwordDigitsDescription       = "Use digits [0-9] when generating the password. Applies to the `random` type."
	passwordSymbolsDescription      = "Use symbols [!@.-_*] when generating the password. Applies to the `random` type."
	passwordExcludeCharsDescription = "Characters to exclude when generating the password. Applies to the `random` type. On 1Password Connect this requires Connect v1.4.0 or later."
	passwordWordCountDescription    = "The number of words in the password. Applies to the `memorable` type."
	passwordSeparatorDescription    = "The separator between words. Applies to the `memorable` type."
	passwordCapitalizeDescription   = "Uppercase one randomly selected word. Applies to the `memorable` type."
	passwordWordListDescription     = "The type of word list used. Applies to the `memorable` type."

	enumDescription = "%s One of %q"

	OTPFieldIDPrefix = "TOTP_"
)

var (
	recipeTypes = []string{"random", "memorable", "pin"}

	recipeSeparators = []string{"digits", "digits_and_symbols", "spaces", "hyphens", "underscores", "periods", "commas"}

	recipeWordLists = []string{"full_words", "syllables", "three_letters"}

	dbTypes = []string{"db2", "filemaker", "msaccess", "mssql", "mysql", "oracle", "postgresql", "sqlite", "other"}

	categories = []string{
		strings.ToLower(string(model.Login)),
		strings.ToLower(string(model.Password)),
		strings.ToLower(string(model.Database)),
		strings.ToLower(string(model.SecureNote)),
		strings.ToLower(string(model.SSHKey)),
		strings.ToLower(string(model.APICredential)),
		strings.ToLower(string(model.Server)),
		strings.ToLower(string(model.Router)),
		strings.ToLower(string(model.SoftwareLicense)),
	}
	dataSourceCategories = append(categories,
		strings.ToLower(string(model.Document)),
	)

	fieldPurposes = []string{
		string(model.FieldPurposeUsername),
		string(model.FieldPurposePassword),
		string(model.FieldPurposeNotes),
	}

	fieldTypes = []string{
		string(model.FieldTypeString),
		string(model.FieldTypeConcealed),
		string(model.FieldTypeEmail),
		string(model.FieldTypeURL),
		string(model.FieldTypeOTP),
		string(model.FieldTypeDate),
		string(model.FieldTypeMonthYear),
		string(model.FieldTypeMenu),
		string(model.FieldTypeSSHKey),
	}

	sshKeyTypes = []string{"ed25519", "rsa"}
)
