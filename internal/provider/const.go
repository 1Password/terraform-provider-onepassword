package provider

const (
	itemUUIDDescription  = "The UUID of the item. Item identifiers are unique within a specific vault."
	vaultUUIDDescription = "The UUID of the vault the item is in."
	categoryDescription  = "The category of the item."
	itemTitleDescription = "The title of the item."
	urlDescription       = "The primary URL for the item."
	tagsDescription      = "An array of strings of the tags assigned to the item."
	usernameDescription  = "Username for this item."
	passwordDescription  = "Password for this item."
	noteValueDescription = "Secure Note value."

	dbHostnameDescription = "(Only applies to the database category) The address where the database can be found"
	dbDatabaseDescription = "(Only applies to the database category) The name of the database."
	dbPortDescription     = "(Only applies to the database category) The port the database is listening on."
	dbTypeDescription     = "(Only applies to the database category) The type of database."

	sectionsDescription      = "A list of custom sections in an item"
	sectionDescription       = "A custom section in an item that contains custom fields"
	sectionIDDescription     = "A unique identifier for the section."
	sectionLabelDescription  = "The label for the section."
	sectionFieldsDescription = "A list of custom fields in the section."

	fieldDescription        = "A custom field."
	fieldIDDescription      = "A unique identifier for the field."
	fieldLabelDescription   = "The label for the field."
	fieldPurposeDescription = "Purpose indicates this is a special field: a username, password, or notes field."
	fieldTypeDescription    = "The type of value stored in the field."
	fieldValueDescription   = "The value of the field."

	passwordRecipeDescription  = "The recipe used to generate a new value for a password."
	passwordElementDescription = "The kinds of characters to include in the password."
	passwordLengthDescription  = "The length of the password to be generated."
	passwordLettersDescription = "Use letters [a-zA-Z] when generating the password."
	passwordDigitsDescription  = "Use digits [0-9] when generating the password."
	passwordSymbolsDescription = "Use symbols [!@.-_*] when generating the password."

	enumDescription = "%s One of %q"
)

var (
	categories    = []string{"login", "password", "database"}
	dbTypes       = []string{"db2", "filemaker", "msaccess", "mssql", "mysql", "oracle", "postgresql", "sqlite", "other"}
	fieldPurposes = []string{"USERNAME", "PASSWORD", "NOTES"}
	fieldTypes    = []string{"STRING", "EMAIL", "CONCEALED", "URL", "OTP", "DATE", "MONTH_YEAR", "MENU"}
)