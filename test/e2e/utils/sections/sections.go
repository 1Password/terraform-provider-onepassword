package sections

type TestField struct {
	Label string
	Value string
	Type  string
}

type TestSection struct {
	Label  string
	Fields []TestField
}

type TestSectionData struct {
	Sections []TestSection
}

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

func mapFields(fields []TestField) []map[string]any {
	var mappedFields []map[string]any

	for _, f := range fields {
		mappedFields = append(mappedFields, map[string]any{
			"label": f.Label,
			"value": f.Value,
			"type":  f.Type,
		})
	}

	return mappedFields
}
