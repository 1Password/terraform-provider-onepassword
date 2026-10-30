package sections

type TestField struct {
	Label          string
	Value          string
	Type           string
	PasswordRecipe *map[string]any
	// ValueWO and ValueWOVersion set the field's write-only value; both are left out when zero.
	ValueWO        string
	ValueWOVersion int
}

type TestSection struct {
	Label  string
	Fields []TestField
}

type TestSectionData struct {
	Sections []TestSection
}

// TestSectionMapField represents a field in section_map
type TestSectionMapField struct {
	Type           string
	Value          string
	PasswordRecipe *map[string]any
	// ValueWO and ValueWOVersion set the field's write-only value; both are left out when zero.
	ValueWO        string
	ValueWOVersion int
}

// TestSectionMapEntry represents a section entry in section_map
type TestSectionMapEntry struct {
	FieldMap map[string]TestSectionMapField
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

		if f.ValueWO != "" {
			fieldMap["value_wo"] = f.ValueWO
		}

		if f.ValueWOVersion != 0 {
			fieldMap["value_wo_version"] = f.ValueWOVersion
		}

		mappedFields = append(mappedFields, fieldMap)
	}
	return mappedFields
}

// BuildSectionMap converts a map of section label to TestSectionMapEntry
func BuildSectionMap(sections map[string]TestSectionMapEntry) map[string]any {
	result := make(map[string]any)

	for sectionLabel, sectionEntry := range sections {
		fieldMap := make(map[string]any)

		for fieldLabel, field := range sectionEntry.FieldMap {
			fieldAttrs := map[string]any{
				"type": field.Type,
			}

			if field.Value != "" {
				fieldAttrs["value"] = field.Value
			}

			if field.ValueWO != "" {
				fieldAttrs["value_wo"] = field.ValueWO
			}

			if field.ValueWOVersion != 0 {
				fieldAttrs["value_wo_version"] = field.ValueWOVersion
			}

			if field.PasswordRecipe != nil && len(*field.PasswordRecipe) > 0 {
				fieldAttrs["password_recipe"] = *field.PasswordRecipe
			}

			fieldMap[fieldLabel] = fieldAttrs
		}

		result[sectionLabel] = map[string]any{
			"field_map": fieldMap,
		}
	}

	return result
}
