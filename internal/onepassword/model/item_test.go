package model

import (
	"reflect"
	"testing"

	sdk "github.com/1password/onepassword-sdk-go"
)

func stringPtr(s string) *string {
	return &s
}

func TestFromModelCategoryToSDK(t *testing.T) {
	tests := map[string]struct {
		input    ItemCategory
		expected sdk.ItemCategory
	}{
		"should convert Login category": {
			input:    Login,
			expected: sdk.ItemCategoryLogin,
		},
		"should convert Password category": {
			input:    Password,
			expected: sdk.ItemCategoryPassword,
		},
		"should convert SecureNote category": {
			input:    SecureNote,
			expected: sdk.ItemCategorySecureNote,
		},
		"should convert Document category": {
			input:    Document,
			expected: sdk.ItemCategoryDocument,
		},
		"should convert SSHKey category": {
			input:    SSHKey,
			expected: sdk.ItemCategorySSHKey,
		},
		"should convert Database category": {
			input:    Database,
			expected: sdk.ItemCategoryDatabase,
		},
		"should return zero value for unknown category": {
			input:    ItemCategory("UNKNOWN"),
			expected: sdk.ItemCategory(""),
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromModelCategoryToSDK(test.input)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestFromSDKCategoryToModel(t *testing.T) {
	tests := map[string]struct {
		input    sdk.ItemCategory
		expected ItemCategory
	}{
		"should convert Login category": {
			input:    sdk.ItemCategoryLogin,
			expected: Login,
		},
		"should convert Password category": {
			input:    sdk.ItemCategoryPassword,
			expected: Password,
		},
		"should convert SecureNote category": {
			input:    sdk.ItemCategorySecureNote,
			expected: SecureNote,
		},
		"should convert Document category": {
			input:    sdk.ItemCategoryDocument,
			expected: Document,
		},
		"should convert SSHKey category": {
			input:    sdk.ItemCategorySSHKey,
			expected: SSHKey,
		},
		"should convert Database category": {
			input:    sdk.ItemCategoryDatabase,
			expected: Database,
		},
		"should return zero value for unknown category": {
			input:    sdk.ItemCategory("UNKNOWN"),
			expected: ItemCategory(""),
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromSDKCategoryToModel(test.input)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestToModelFieldType(t *testing.T) {
	tests := map[string]struct {
		input    sdk.ItemFieldType
		expected ItemFieldType
	}{
		"should convert Concealed field type": {
			input:    sdk.ItemFieldTypeConcealed,
			expected: FieldTypeConcealed,
		},
		"should convert Date field type": {
			input:    sdk.ItemFieldTypeDate,
			expected: FieldTypeDate,
		},
		"should convert Email field type": {
			input:    sdk.ItemFieldTypeEmail,
			expected: FieldTypeEmail,
		},
		"should convert Menu field type": {
			input:    sdk.ItemFieldTypeMenu,
			expected: FieldTypeMenu,
		},
		"should convert MonthYear field type": {
			input:    sdk.ItemFieldTypeMonthYear,
			expected: FieldTypeMonthYear,
		},
		"should convert TOTP field type": {
			input:    sdk.ItemFieldTypeTOTP,
			expected: FieldTypeOTP,
		},
		"should convert Text field type to String": {
			input:    sdk.ItemFieldTypeText,
			expected: FieldTypeString,
		},
		"should convert URL field type": {
			input:    sdk.ItemFieldTypeURL,
			expected: FieldTypeURL,
		},
		"should return zero value for unknown field type": {
			input:    sdk.ItemFieldType("UNKNOWN"),
			expected: ItemFieldType(""),
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toModelFieldType(test.input)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestToSDKFieldType(t *testing.T) {
	tests := map[string]struct {
		input    ItemFieldType
		expected sdk.ItemFieldType
	}{
		"should convert Concealed field type": {
			input:    FieldTypeConcealed,
			expected: sdk.ItemFieldTypeConcealed,
		},
		"should convert Date field type": {
			input:    FieldTypeDate,
			expected: sdk.ItemFieldTypeDate,
		},
		"should convert Email field type": {
			input:    FieldTypeEmail,
			expected: sdk.ItemFieldTypeEmail,
		},
		"should convert Menu field type": {
			input:    FieldTypeMenu,
			expected: sdk.ItemFieldTypeMenu,
		},
		"should convert MonthYear field type": {
			input:    FieldTypeMonthYear,
			expected: sdk.ItemFieldTypeMonthYear,
		},
		"should convert OTP field type to TOTP": {
			input:    FieldTypeOTP,
			expected: sdk.ItemFieldTypeTOTP,
		},
		"should convert String field type to Text": {
			input:    FieldTypeString,
			expected: sdk.ItemFieldTypeText,
		},
		"should convert URL field type": {
			input:    FieldTypeURL,
			expected: sdk.ItemFieldTypeURL,
		},
		"should return zero value for unknown field type": {
			input:    ItemFieldType("UNKNOWN"),
			expected: sdk.ItemFieldType(""),
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toSDKFieldType(test.input)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

// Test bidirectional conversions to ensure round-trip consistency
func TestCategoryConversionRoundTrip(t *testing.T) {
	categories := []ItemCategory{
		Login,
		Password,
		SecureNote,
		Document,
		SSHKey,
		Database,
	}

	for _, category := range categories {
		t.Run(string(category), func(t *testing.T) {
			sdkCategory := fromModelCategoryToSDK(category)
			modelCategory := fromSDKCategoryToModel(sdkCategory)
			if modelCategory != category {
				t.Errorf("Round-trip conversion failed: %v -> %v -> %v", category, sdkCategory, modelCategory)
			}
		})
	}
}

func TestFieldTypeConversionRoundTrip(t *testing.T) {
	fieldTypes := []ItemFieldType{
		FieldTypeConcealed,
		FieldTypeDate,
		FieldTypeEmail,
		FieldTypeMenu,
		FieldTypeMonthYear,
		FieldTypeOTP,
		FieldTypeString,
		FieldTypeURL,
	}

	for _, fieldType := range fieldTypes {
		t.Run(string(fieldType), func(t *testing.T) {
			sdkFieldType := toSDKFieldType(fieldType)
			modelFieldType := toModelFieldType(sdkFieldType)
			if modelFieldType != fieldType {
				t.Errorf("Round-trip conversion failed: %v -> %v -> %v", fieldType, sdkFieldType, modelFieldType)
			}
		})
	}
}

func TestFromSDKURLs(t *testing.T) {
	tests := map[string]struct {
		input    []sdk.Website
		expected []ItemURL
	}{
		"should convert empty slice": {
			input:    []sdk.Website{},
			expected: []ItemURL{},
		},
		"should convert single website": {
			input: []sdk.Website{
				{URL: "https://example.com", Label: "Example"},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
			},
		},
		"should convert multiple websites with first as primary": {
			input: []sdk.Website{
				{URL: "https://example.com", Label: "Example"},
				{URL: "https://test.com", Label: "Test"},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
				{URL: "https://test.com", Label: "Test", Primary: false},
			},
		},
		"should handle empty label": {
			input: []sdk.Website{
				{URL: "https://example.com", Label: ""},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "", Primary: true},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromSDKURLs(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromSDKSections(t *testing.T) {
	tests := map[string]struct {
		input    map[string]ItemSection
		expected []ItemSection
	}{
		"should convert empty map": {
			input:    map[string]ItemSection{},
			expected: []ItemSection{},
		},
		"should convert single section": {
			input: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
			expected: []ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
		},
		"should convert multiple sections": {
			input: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
				"section2": {ID: "section2", Label: "Section 2"},
			},
			expected: []ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromSDKSections(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %d sections, got %d", len(test.expected), len(actual))
			}
			// Check that all expected sections are present
			for _, expected := range test.expected {
				found := false
				for _, actualSection := range actual {
					if actualSection.ID == expected.ID && actualSection.Label == expected.Label {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected section %+v not found in result", expected)
				}
			}
		})
	}
}

func TestBuildSectionMap(t *testing.T) {
	tests := map[string]struct {
		input    *sdk.Item
		expected map[string]ItemSection
	}{
		"should build empty map from item with no sections": {
			input: &sdk.Item{
				Sections: []sdk.ItemSection{},
			},
			expected: map[string]ItemSection{},
		},
		"should build map with single section": {
			input: &sdk.Item{
				Sections: []sdk.ItemSection{
					{ID: "section1", Title: "Section 1"},
				},
			},
			expected: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
		},
		"should build map with multiple sections": {
			input: &sdk.Item{
				Sections: []sdk.ItemSection{
					{ID: "section1", Title: "Section 1"},
					{ID: "section2", Title: "Section 2"},
				},
			},
			expected: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
				"section2": {ID: "section2", Label: "Section 2"},
			},
		},
		"should skip sections with empty ID": {
			input: &sdk.Item{
				Sections: []sdk.ItemSection{
					{ID: "section1", Title: "Section 1"},
					{ID: "", Title: "Empty ID"},
				},
			},
			expected: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := buildSectionMap(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromSDKFields(t *testing.T) {
	sectionID := "section1"
	sectionMap := map[string]ItemSection{
		"section1": {ID: "section1", Label: "Section 1"},
	}

	tests := map[string]struct {
		input    *sdk.Item
		expected []ItemField
	}{
		"should convert empty fields": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{},
			},
			expected: []ItemField{},
		},
		"should convert basic field": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
			},
		},
		"should set username purpose": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "username",
						Title:     "Username",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "user1",
					},
				},
			},
			expected: []ItemField{
				{
					ID:      "username",
					Label:   "Username",
					Type:    FieldTypeString,
					Value:   "user1",
					Purpose: FieldPurposeUsername,
				},
			},
		},
		"should set password purpose": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "password",
						Title:     "Password",
						FieldType: sdk.ItemFieldTypeConcealed,
						Value:     "secret",
					},
				},
			},
			expected: []ItemField{
				{
					ID:      "password",
					Label:   "Password",
					Type:    FieldTypeConcealed,
					Value:   "secret",
					Purpose: FieldPurposePassword,
				},
			},
		},
		"should associate field with section": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: &sectionID,
					},
				},
			},
			expected: []ItemField{
				{
					ID:           "field1",
					Label:        "Field 1",
					Type:         FieldTypeString,
					Value:        "value1",
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
		},
		"should handle field with unknown section": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: stringPtr("unknown"),
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
			},
		},
		"should handle field with empty section ID": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: stringPtr(""),
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
			},
		},
		"should handle field with nil section ID": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: nil,
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromSDKFields(test.input, sectionMap)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %d fields, got %d", len(test.expected), len(actual))
				return
			}
			for i, expected := range test.expected {
				if i >= len(actual) {
					t.Errorf("Missing field at index %d", i)
					continue
				}
				actualField := actual[i]
				if !reflect.DeepEqual(actualField, expected) {
					t.Errorf("Field mismatch at index %d: expected %+v, got %+v", i, expected, actualField)
				}
			}
		})
	}
}

func TestFromSDKFiles(t *testing.T) {
	sectionID := "section1"
	sectionMap := map[string]ItemSection{
		"section1": {ID: "section1", Label: "Section 1"},
	}

	tests := map[string]struct {
		input    *sdk.Item
		expected []ItemFile
	}{
		"should convert empty files": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{},
			},
			expected: []ItemFile{},
		},
		"should convert single file": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file.txt",
							Size: 1024,
						},
					},
				},
			},
			expected: []ItemFile{
				{
					ID:   "file1",
					Name: "file.txt",
					Size: 1024,
				},
			},
		},
		"should associate file with section": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file.txt",
							Size: 1024,
						},
						SectionID: sectionID,
					},
				},
			},
			expected: []ItemFile{
				{
					ID:           "file1",
					Name:         "file.txt",
					Size:         1024,
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
		},
		"should handle file with unknown section": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file.txt",
							Size: 1024,
						},
						SectionID: "unknown",
					},
				},
			},
			expected: []ItemFile{
				{
					ID:   "file1",
					Name: "file.txt",
					Size: 1024,
				},
			},
		},
		"should append document if present": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file.txt",
							Size: 1024,
						},
					},
				},
				Document: &sdk.FileAttributes{
					ID:   "doc1",
					Name: "document.pdf",
					Size: 2048,
				},
			},
			expected: []ItemFile{
				{
					ID:   "file1",
					Name: "file.txt",
					Size: 1024,
				},
				{
					ID:   "doc1",
					Name: "document.pdf",
					Size: 2048,
				},
			},
		},
		"should handle multiple files with document": {
			input: &sdk.Item{
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file1.txt",
							Size: 1024,
						},
					},
					{
						Attributes: sdk.FileAttributes{
							ID:   "file2",
							Name: "file2.txt",
							Size: 512,
						},
					},
				},
				Document: &sdk.FileAttributes{
					ID:   "doc1",
					Name: "document.pdf",
					Size: 2048,
				},
			},
			expected: []ItemFile{
				{
					ID:   "file1",
					Name: "file1.txt",
					Size: 1024,
				},
				{
					ID:   "file2",
					Name: "file2.txt",
					Size: 512,
				},
				{
					ID:   "doc1",
					Name: "document.pdf",
					Size: 2048,
				},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromSDKFiles(test.input, sectionMap)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %d files, got %d", len(test.expected), len(actual))
				return
			}
			for i, expected := range test.expected {
				if i >= len(actual) {
					t.Errorf("Missing file at index %d", i)
					continue
				}
				actualFile := actual[i]
				if !reflect.DeepEqual(actualFile, expected) {
					t.Errorf("File mismatch at index %d: expected %+v, got %+v", i, expected, actualFile)
				}
			}
		})
	}
}

func TestToSDKFields(t *testing.T) {
	notesValue := "These are notes"
	tests := map[string]struct {
		inputFields    []ItemField
		expectedFields []sdk.ItemField
		expectedNotes  *string
	}{
		"should convert empty fields": {
			inputFields:    []ItemField{},
			expectedFields: []sdk.ItemField{},
			expectedNotes:  nil,
		},
		"should convert single field": {
			inputFields: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
			},
			expectedFields: []sdk.ItemField{
				{
					ID:        "field1",
					Title:     "Field 1",
					FieldType: sdk.ItemFieldTypeText,
					Value:     "value1",
				},
			},
			expectedNotes: nil,
		},
		"should extract notes field": {
			inputFields: []ItemField{
				{
					ID:      "notes",
					Label:   "Notes",
					Type:    FieldTypeString,
					Purpose: FieldPurposeNotes,
					Value:   notesValue,
				},
			},
			expectedFields: []sdk.ItemField{},
			expectedNotes:  &notesValue,
		},
		"should convert fields and extract notes": {
			inputFields: []ItemField{
				{
					ID:    "field1",
					Label: "Field 1",
					Type:  FieldTypeString,
					Value: "value1",
				},
				{
					ID:      "notes",
					Label:   "Notes",
					Type:    FieldTypeString,
					Purpose: FieldPurposeNotes,
					Value:   notesValue,
				},
				{
					ID:    "field2",
					Label: "Field 2",
					Type:  FieldTypeEmail,
					Value: "test@example.com",
				},
			},
			expectedFields: []sdk.ItemField{
				{
					ID:        "field1",
					Title:     "Field 1",
					FieldType: sdk.ItemFieldTypeText,
					Value:     "value1",
				},
				{
					ID:        "field2",
					Title:     "Field 2",
					FieldType: sdk.ItemFieldTypeEmail,
					Value:     "test@example.com",
				},
			},
			expectedNotes: &notesValue,
		},
		"should handle field with section ID": {
			inputFields: []ItemField{
				{
					ID:        "field1",
					Label:     "Field 1",
					Type:      FieldTypeString,
					Value:     "value1",
					SectionID: "section1",
				},
			},
			expectedFields: []sdk.ItemField{
				{
					ID:        "field1",
					Title:     "Field 1",
					FieldType: sdk.ItemFieldTypeText,
					Value:     "value1",
					SectionID: stringPtr("section1"),
				},
			},
			expectedNotes: nil,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actualFields, actualNotes := toSDKFields(test.inputFields)
			if len(actualFields) != len(test.expectedFields) {
				t.Errorf("Expected %d fields, got %d", len(test.expectedFields), len(actualFields))
				return
			}
			for i, expected := range test.expectedFields {
				if i >= len(actualFields) {
					t.Errorf("Missing field at index %d", i)
					continue
				}
				actualField := actualFields[i]
				if !reflect.DeepEqual(actualField, expected) {
					t.Errorf("Field mismatch at index %d: expected %+v, got %+v", i, expected, actualField)
				}
			}
			if (test.expectedNotes == nil) != (actualNotes == nil) {
				t.Errorf("Notes mismatch: expected %v, got %v", test.expectedNotes, actualNotes)
			} else if test.expectedNotes != nil && actualNotes != nil && *test.expectedNotes != *actualNotes {
				t.Errorf("Notes value mismatch: expected %s, got %s", *test.expectedNotes, *actualNotes)
			}
		})
	}
}

func TestToSDKField(t *testing.T) {
	tests := map[string]struct {
		input    ItemField
		expected sdk.ItemField
	}{
		"should convert basic field": {
			input: ItemField{
				ID:    "field1",
				Label: "Field 1",
				Type:  FieldTypeString,
				Value: "value1",
			},
			expected: sdk.ItemField{
				ID:        "field1",
				Title:     "Field 1",
				FieldType: sdk.ItemFieldTypeText,
				Value:     "value1",
			},
		},
		"should convert field with section ID": {
			input: ItemField{
				ID:        "field1",
				Label:     "Field 1",
				Type:      FieldTypeString,
				Value:     "value1",
				SectionID: "section1",
			},
			expected: sdk.ItemField{
				ID:        "field1",
				Title:     "Field 1",
				FieldType: sdk.ItemFieldTypeText,
				Value:     "value1",
				SectionID: stringPtr("section1"),
			},
		},
		"should convert field without section ID": {
			input: ItemField{
				ID:    "field1",
				Label: "Field 1",
				Type:  FieldTypeString,
				Value: "value1",
			},
			expected: sdk.ItemField{
				ID:        "field1",
				Title:     "Field 1",
				FieldType: sdk.ItemFieldTypeText,
				Value:     "value1",
			},
		},
		"should convert different field types": {
			input: ItemField{
				ID:    "field1",
				Label: "Field 1",
				Type:  FieldTypeEmail,
				Value: "test@example.com",
			},
			expected: sdk.ItemField{
				ID:        "field1",
				Title:     "Field 1",
				FieldType: sdk.ItemFieldTypeEmail,
				Value:     "test@example.com",
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toSDKField(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Field mismatch: expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestToSDKSections(t *testing.T) {
	tests := map[string]struct {
		input    []ItemSection
		expected []sdk.ItemSection
	}{
		"should convert empty sections": {
			input:    []ItemSection{},
			expected: []sdk.ItemSection{},
		},
		"should convert single section": {
			input: []ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			expected: []sdk.ItemSection{
				{ID: "section1", Title: "Section 1"},
			},
		},
		"should convert multiple sections": {
			input: []ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			expected: []sdk.ItemSection{
				{ID: "section1", Title: "Section 1"},
				{ID: "section2", Title: "Section 2"},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toSDKSections(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %d sections, got %d", len(test.expected), len(actual))
				return
			}
			for i, expected := range test.expected {
				if i >= len(actual) {
					t.Errorf("Missing section at index %d", i)
					continue
				}
				actualSection := actual[i]
				if !reflect.DeepEqual(actualSection, expected) {
					t.Errorf("Section mismatch at index %d: expected %+v, got %+v", i, expected, actualSection)
				}
			}
		})
	}
}

func TestToSDKWebsites(t *testing.T) {
	tests := map[string]struct {
		input    []ItemURL
		expected []sdk.Website
	}{
		"should convert empty URLs": {
			input:    []ItemURL{},
			expected: []sdk.Website{},
		},
		"should convert single URL": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "Example"},
			},
			expected: []sdk.Website{
				{URL: "https://example.com", Label: "Example"},
			},
		},
		"should convert multiple URLs": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "Example"},
				{URL: "https://test.com", Label: "Test"},
			},
			expected: []sdk.Website{
				{URL: "https://example.com", Label: "Example"},
				{URL: "https://test.com", Label: "Test"},
			},
		},
		"should skip URLs with empty URL": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "Example"},
				{URL: "", Label: "Empty"},
			},
			expected: []sdk.Website{
				{URL: "https://example.com", Label: "Example"},
			},
		},
		"should handle empty label": {
			input: []ItemURL{
				{URL: "https://example.com", Label: ""},
			},
			expected: []sdk.Website{
				{URL: "https://example.com", Label: ""},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toSDKWebsites(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %d websites, got %d", len(test.expected), len(actual))
				return
			}
			for i, expected := range test.expected {
				if i >= len(actual) {
					t.Errorf("Missing website at index %d", i)
					continue
				}
				actualWebsite := actual[i]
				if !reflect.DeepEqual(actualWebsite, expected) {
					t.Errorf("Website mismatch at index %d: expected %+v, got %+v", i, expected, actualWebsite)
				}
			}
		})
	}
}

func TestFromSDKItemToModel(t *testing.T) {
	sectionID := "section1"
	notesValue := "Test notes"

	tests := map[string]struct {
		input    *sdk.Item
		expected *Item
		wantErr  bool
	}{
		"should return error for nil item": {
			input:   nil,
			wantErr: true,
		},
		"should convert basic item": {
			input: &sdk.Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: sdk.ItemCategoryLogin,
				Tags:     []string{"tag1", "tag2"},
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				Tags:     []string{"tag1", "tag2"},
				URLs:     []ItemURL{},
				Sections: []ItemSection{},
				Fields:   []ItemField{},
				Files:    []ItemFile{},
			},
			wantErr: false,
		},
		"should convert item with notes": {
			input: &sdk.Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: sdk.ItemCategoryLogin,
				Notes:    notesValue,
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				URLs:     []ItemURL{},
				Sections: []ItemSection{},
				Fields: []ItemField{
					{
						Type:    FieldTypeString,
						Purpose: FieldPurposeNotes,
						Value:   notesValue,
					},
				},
				Files: []ItemFile{},
			},
			wantErr: false,
		},
		"should convert item with sections and fields": {
			input: &sdk.Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{
					{ID: "section1", Title: "Section 1"},
				},
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: &sectionID,
					},
				},
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				URLs:     []ItemURL{},
				Sections: []ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []ItemField{
					{
						ID:           "field1",
						Label:        "Field 1",
						Type:         FieldTypeString,
						Value:        "value1",
						SectionID:    "section1",
						SectionLabel: "Section 1",
					},
				},
				Files: []ItemFile{},
			},
			wantErr: false,
		},
		"should convert item with websites": {
			input: &sdk.Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: sdk.ItemCategoryLogin,
				Websites: []sdk.Website{
					{URL: "https://example.com", Label: "Example"},
					{URL: "https://test.com", Label: "Test"},
				},
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				URLs: []ItemURL{
					{URL: "https://example.com", Label: "Example", Primary: true},
					{URL: "https://test.com", Label: "Test", Primary: false},
				},
				Sections: []ItemSection{},
				Fields:   []ItemField{},
				Files:    []ItemFile{},
			},
			wantErr: false,
		},
		"should convert item with files": {
			input: &sdk.Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: sdk.ItemCategoryLogin,
				Files: []sdk.ItemFile{
					{
						Attributes: sdk.FileAttributes{
							ID:   "file1",
							Name: "file.txt",
							Size: 1024,
						},
					},
				},
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				URLs:     []ItemURL{},
				Sections: []ItemSection{},
				Fields:   []ItemField{},
				Files: []ItemFile{
					{
						ID:   "file1",
						Name: "file.txt",
						Size: 1024,
					},
				},
			},
			wantErr: false,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			item := &Item{}
			err := item.FromSDKItemToModel(test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("FromSDKItemToModel() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(item, test.expected) {
				t.Errorf("Item mismatch: got %+v, expected %+v", item, test.expected)
			}
		})
	}
}

func TestFromModelItemToSDKCreateParams(t *testing.T) {
	notesValue := "Test notes"
	sectionID := "section1"

	tests := map[string]struct {
		input    *Item
		expected sdk.ItemCreateParams
	}{
		"should convert basic item": {
			input: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				Tags:     []string{"tag1", "tag2"},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Tags:     []string{"tag1", "tag2"},
				Sections: []sdk.ItemSection{},
				Websites: []sdk.Website{},
				Fields:   []sdk.ItemField{},
				Files:    []sdk.FileCreateParams{},
			},
		},
		"should convert item with sections": {
			input: &Item{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: Login,
				Sections: []ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{
					{ID: "section1", Title: "Section 1"},
				},
				Websites: []sdk.Website{},
				Fields:   []sdk.ItemField{},
				Files:    []sdk.FileCreateParams{},
			},
		},
		"should convert item with URLs": {
			input: &Item{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: Login,
				URLs: []ItemURL{
					{URL: "https://example.com", Label: "Example"},
					{URL: "https://test.com", Label: "Test"},
				},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{},
				Websites: []sdk.Website{
					{URL: "https://example.com", Label: "Example"},
					{URL: "https://test.com", Label: "Test"},
				},
				Fields: []sdk.ItemField{},
				Files:  []sdk.FileCreateParams{},
			},
		},
		"should convert item with fields and notes": {
			input: &Item{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: Login,
				Fields: []ItemField{
					{
						ID:    "field1",
						Label: "Field 1",
						Type:  FieldTypeString,
						Value: "value1",
					},
					{
						ID:      "notes",
						Label:   "Notes",
						Type:    FieldTypeString,
						Purpose: FieldPurposeNotes,
						Value:   notesValue,
					},
				},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{},
				Websites: []sdk.Website{},
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
					},
				},
				Notes: &notesValue,
				Files: []sdk.FileCreateParams{},
			},
		},
		"should convert item with field section ID": {
			input: &Item{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: Login,
				Fields: []ItemField{
					{
						ID:        "field1",
						Label:     "Field 1",
						Type:      FieldTypeString,
						Value:     "value1",
						SectionID: sectionID,
					},
				},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{},
				Websites: []sdk.Website{},
				Fields: []sdk.ItemField{
					{
						ID:        "field1",
						Title:     "Field 1",
						FieldType: sdk.ItemFieldTypeText,
						Value:     "value1",
						SectionID: &sectionID,
					},
				},
				Files: []sdk.FileCreateParams{},
			},
		},
		"should skip URLs with empty URL": {
			input: &Item{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: Login,
				URLs: []ItemURL{
					{URL: "https://example.com", Label: "Example"},
					{URL: "", Label: "Empty"},
				},
			},
			expected: sdk.ItemCreateParams{
				VaultID:  "vault1",
				Title:    "Test Item",
				Category: sdk.ItemCategoryLogin,
				Sections: []sdk.ItemSection{},
				Websites: []sdk.Website{
					{URL: "https://example.com", Label: "Example"},
				},
				Fields: []sdk.ItemField{},
				Files:  []sdk.FileCreateParams{},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := test.input.FromModelItemToSDKCreateParams()
			if actual.VaultID != test.expected.VaultID ||
				actual.Title != test.expected.Title ||
				actual.Category != test.expected.Category {
				t.Errorf("Basic fields mismatch: got %+v, expected %+v", actual, test.expected)
			}
			if !reflect.DeepEqual(actual.Tags, test.expected.Tags) {
				t.Errorf("Tags mismatch: got %+v, expected %+v", actual.Tags, test.expected.Tags)
			}
			if !reflect.DeepEqual(actual.Sections, test.expected.Sections) {
				t.Errorf("Sections mismatch: got %+v, expected %+v", actual.Sections, test.expected.Sections)
			}
			if !reflect.DeepEqual(actual.Fields, test.expected.Fields) {
				t.Errorf("Fields mismatch: got %+v, expected %+v", actual.Fields, test.expected.Fields)
			}
			if !reflect.DeepEqual(actual.Websites, test.expected.Websites) {
				t.Errorf("Websites mismatch: got %+v, expected %+v", actual.Websites, test.expected.Websites)
			}
			if !reflect.DeepEqual(actual.Notes, test.expected.Notes) {
				t.Errorf("Notes mismatch: got %+v, expected %+v", actual.Notes, test.expected.Notes)
			}
		})
	}
}
