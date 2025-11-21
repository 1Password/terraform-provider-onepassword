package sections

type TestField struct {
	Label          string
	Value          string
	Type           string
	PasswordRecipe *map[string]any
}

type TestSection struct {
	Label  string
	Fields []TestField
}

type TestSectionData struct {
	Sections []TestSection
}

// MapSections converts a list of TestSection to a list of maps that can be used in Terraform configuration
func MapSections(sections []TestSection) []map[string]any {
	mappedSections := make([]map[string]any, len(sections))

	for i, s := range sections {
		sectionMap := map[string]any{
			"label": s.Label,
		}

		if len(s.Fields) > 0 {
			sectionMap["field"] = mapFields(s.Fields)
		}

		mappedSections[i] = sectionMap
	}
	return mappedSections
}

// mapFields converts a list of TestField to a list of maps that can be used in Terraform configuration
func mapFields(fields []TestField) []map[string]any {
	var mappedFields []map[string]any
	for _, f := range fields {
		fieldMap := map[string]any{
			"label": f.Label,
			"type":  f.Type,
		}

		if f.PasswordRecipe != nil && len(*f.PasswordRecipe) > 0 {
			fieldMap["password_recipe"] = []map[string]any{*f.PasswordRecipe}
		}

		if f.Value != "" {
			fieldMap["value"] = f.Value
		}

		mappedFields = append(mappedFields, fieldMap)
	}
	return mappedFields
}
