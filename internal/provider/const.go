package provider

import (
	"strings"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword/model"
)

const (
	terraformItemIDDescription = "The Terraform resource identifier for this item in the format `vaults/<vault_id>/items/<item_id>`."

	itemUUIDDescription                 = "The UUID of the item. Item identifiers are unique within a specific vault."
	vaultUUIDDescription                = "The UUID of the vault the item is in."
	categoryDescription                 = "The category of the item."
	itemTitleDescription                = "The title of the item."
	urlDescription                      = "The primary URL for the item."
	tagsDescription                     = "An array of strings of the tags assigned to the item."
	usernameDescription                 = "Username for this item."
	passwordDescription                 = "Password for this item."
	passwordWriteOnceDescription        = "A write-only password for this item. This value is not stored in the state and is intended for use with ephemeral values. **Note**: Write-only arguments require Terraform 1.11 or later."
	passwordWriteOnceVersionDescription = "An integer that must be incremented to trigger an update to the 'password_wo' field."
	credentialDescription               = "API credential for this item."
	noteValueDescription                = "Secure Note value."
	publicKeyDescription                = "SSH Public Key for this item."
	privateKeyDescription               = "SSH Private Key in PKCS#8 for this item."
	privateKeyOpenSSHDescription        = "SSH Private key in OpenSSH format."

	dbHostnameDescription = "(Only applies to the database category) The address where the database can be found"
	dbDatabaseDescription = "(Only applies to the database category) The name of the database."
	dbPortDescription     = "(Only applies to the database category) The port the database is listening on."
	dbTypeDescription     = "(Only applies to the database category) The type of database."

	sectionsDescription            = "A list of custom sections in an item"
	sectionListDescriptionResource = "A list of custom sections in an item. Cannot be used together with `section_map`. Use either `section` (list) or `section_map` (map), but not both."
	sectionMapDescriptionResource  = "A map of custom sections in an item, keyed by section label. This allows direct lookup of sections and their fields by label. Cannot be used together with `section`. Use either `section` (list) or `section_map` (map), but not both."
	sectionIDDescription           = "A unique identifier for the section."
	sectionLabelDescription        = "The label for the section."
	sectionFieldsListDescription   = "A list of custom fields in the section."
	sectionFieldsMapDescription    = "A map of custom fields in the section, keyed by field label."
	sectionFilesDescription        = "A list of files attached to the section."

	filesDescription             = "A list of files attached to the item."
	fileIDDescription            = "The UUID of the file."
	fileNameDescription          = "The name of the file."
	fileContentDescription       = "The content of the file."
	fileContentBase64Description = "The content of the file in base64 encoding. (Use this for binary files.)"

	fieldIDDescription    = "A unique identifier for the field."
	fieldLabelDescription = "The label for the field."
	fieldTypeDescription  = "The type of value stored in the field."
	fieldValueDescription = "The value of the field."

	passwordRecipeDescription  = "The recipe used to generate a new value for a password."
	passwordLengthDescription  = "The length of the password to be generated."
	passwordDigitsDescription  = "Use digits [0-9] when generating the password."
	passwordSymbolsDescription = "Use symbols [!@.-_*] when generating the password."

	enumDescription = "%s One of %q"

	OTPFieldIDPrefix = "TOTP_"
)

var (
	dbTypes = []string{"db2", "filemaker", "msaccess", "mssql", "mysql", "oracle", "postgresql", "sqlite", "other"}

	categories = []string{
		strings.ToLower(string(model.Login)),
		strings.ToLower(string(model.Password)),
		strings.ToLower(string(model.Database)),
		strings.ToLower(string(model.SecureNote)),
	}
	dataSourceCategories = append(categories,
		strings.ToLower(string(model.Document)),
		strings.ToLower(string(model.SSHKey)),
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
	}
)
