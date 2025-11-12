package model

type ItemCategory string
type ItemFieldType string
type FieldPurpose string

const (
	ItemCategoryLogin         ItemCategory = "Login"
	ItemCategoryPassword      ItemCategory = "Password"
	ItemCategoryAPICredential ItemCategory = "ApiCredential"
	ItemCategoryDatabase      ItemCategory = "Database"
	ItemCategorySecureNote    ItemCategory = "Secure_Note"
	ItemCategorySSHKey        ItemCategory = "Ssh_Key"
	ItemCategoryDocument      ItemCategory = "Document"

	FieldTypeString    ItemFieldType = "Text"
	FieldTypeConcealed ItemFieldType = "Concealed"
	FieldTypeEmail     ItemFieldType = "Email"
	FieldTypeURL       ItemFieldType = "Url"
	FieldTypeDate      ItemFieldType = "Date"
	FieldTypeMenu      ItemFieldType = "Menu"
	FieldTypeSSHKey    ItemFieldType = "Ssh_Key"

	FieldPurposeUsername FieldPurpose = "USERNAME"
	FieldPurposePassword FieldPurpose = "PASSWORD"
	FieldPurposeNotes    FieldPurpose = "NOTES"
)

type Item struct {
	ID       string
	Title    string
	VaultID  string
	Category ItemCategory
	Version  int
	Tags     []string
	URLs     []ItemURL
	Sections []*ItemSection
	Fields   []*ItemField
	Files    []*ItemFile
}

type ItemSection struct {
	ID    string
	Label string
}

type ItemField struct {
	ID       string
	Label    string
	Type     ItemFieldType
	Value    string
	Purpose  FieldPurpose
	Section  *ItemSection
	Recipe   *GeneratorRecipe
	Generate bool
}

type GeneratorRecipe struct {
	Length        int
	CharacterSets []string
}

type ItemURL struct {
	URL     string
	Label   string
	Primary bool
}

type ItemFile struct {
	ID      string
	Name    string
	Size    int
	Section *ItemSection
	content []byte
}
