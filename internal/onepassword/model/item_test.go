package model

import (
	"reflect"
	"testing"

	connect "github.com/1Password/connect-sdk-go/onepassword"
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
		"should convert month-year from MM/YYYY to YYYYMM": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "expiry",
						Title:     "Expiration",
						FieldType: sdk.ItemFieldTypeMonthYear,
						Value:     "01/2024",
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "expiry",
					Label: "Expiration",
					Type:  FieldTypeMonthYear,
					Value: "202401",
				},
			},
		},
		"should handle empty month-year value from SDK": {
			input: &sdk.Item{
				Fields: []sdk.ItemField{
					{
						ID:        "expiry",
						Title:     "Expiration",
						FieldType: sdk.ItemFieldTypeMonthYear,
						Value:     "",
					},
				},
			},
			expected: []ItemField{
				{
					ID:    "expiry",
					Label: "Expiration",
					Type:  FieldTypeMonthYear,
					Value: "",
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
		"should convert month-year from YYYYMM to MM/YYYY for SDK": {
			inputFields: []ItemField{
				{
					ID:    "expiry",
					Label: "Expiration",
					Type:  FieldTypeMonthYear,
					Value: "202401",
				},
			},
			expectedFields: []sdk.ItemField{
				{
					ID:        "expiry",
					Title:     "Expiration",
					FieldType: sdk.ItemFieldTypeMonthYear,
					Value:     "01/2024",
				},
			},
			expectedNotes: nil,
		},
		"should handle empty month-year value to SDK": {
			inputFields: []ItemField{
				{
					ID:    "expiry",
					Label: "Expiration",
					Type:  FieldTypeMonthYear,
					Value: "",
				},
			},
			expectedFields: []sdk.ItemField{
				{
					ID:        "expiry",
					Title:     "Expiration",
					FieldType: sdk.ItemFieldTypeMonthYear,
					Value:     "",
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

func TestFromConnectURLs(t *testing.T) {
	tests := map[string]struct {
		input    []connect.ItemURL
		expected []ItemURL
	}{
		"should convert single URL": {
			input: []connect.ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
			},
		},
		"should convert multiple URLs": {
			input: []connect.ItemURL{
				{URL: "https://example.com", Label: "Primary", Primary: true},
				{URL: "https://example.org", Label: "Secondary", Primary: false},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "Primary", Primary: true},
				{URL: "https://example.org", Label: "Secondary", Primary: false},
			},
		},
		"should handle empty slice": {
			input:    []connect.ItemURL{},
			expected: []ItemURL{},
		},
		"should convert URLs with empty labels": {
			input: []connect.ItemURL{
				{URL: "https://example.com", Label: "", Primary: true},
			},
			expected: []ItemURL{
				{URL: "https://example.com", Label: "", Primary: true},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromConnectURLs(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromConnectSections(t *testing.T) {
	tests := map[string]struct {
		input       []*connect.ItemSection
		expected    []ItemSection
		expectedMap map[string]ItemSection
	}{
		"should convert single section": {
			input: []*connect.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			expected: []ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			expectedMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
		},
		"should convert multiple sections": {
			input: []*connect.ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			expected: []ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			expectedMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
				"section2": {ID: "section2", Label: "Section 2"},
			},
		},
		"should handle empty slice": {
			input:       []*connect.ItemSection{},
			expected:    []ItemSection{},
			expectedMap: map[string]ItemSection{},
		},
		"should skip nil sections": {
			input: []*connect.ItemSection{
				{ID: "section1", Label: "Section 1"},
				nil,
				{ID: "section2", Label: "Section 2"},
			},
			expected: []ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			expectedMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
				"section2": {ID: "section2", Label: "Section 2"},
			},
		},
		"should handle sections with empty labels": {
			input: []*connect.ItemSection{
				{ID: "section1", Label: ""},
			},
			expected: []ItemSection{
				{ID: "section1", Label: ""},
			},
			expectedMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: ""},
			},
		},
		"should handle sections with empty IDs": {
			input: []*connect.ItemSection{
				{ID: "", Label: "Section 1"},
			},
			expected: []ItemSection{
				{ID: "", Label: "Section 1"},
			},
			expectedMap: map[string]ItemSection{
				"": {ID: "", Label: "Section 1"},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			sectionMap := make(map[string]ItemSection)
			actual := fromConnectSections(test.input, sectionMap)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Sections mismatch: got %+v, expected %+v", actual, test.expected)
			}
			if !reflect.DeepEqual(sectionMap, test.expectedMap) {
				t.Errorf("Section map mismatch: got %+v, expected %+v", sectionMap, test.expectedMap)
			}
		})
	}
}

func TestFromConnectFields(t *testing.T) {
	tests := map[string]struct {
		input         []*connect.ItemField
		sectionMap    map[string]ItemSection
		expected      []ItemField
		expectedError bool
	}{
		"should convert simple field": {
			input: []*connect.ItemField{
				{ID: "field1", Label: "Field 1", Type: connect.FieldTypeString, Value: "value1"},
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemField{
				{ID: "field1", Label: "Field 1", Type: FieldTypeString, Value: "value1"},
			},
		},
		"should convert date field from timestamp to YYYY-MM-DD": {
			input: []*connect.ItemField{
				{ID: "date1", Type: connect.FieldTypeDate, Value: "1609459200"}, // 2021-01-01
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemField{
				{ID: "date1", Type: FieldTypeDate, Value: "2021-01-01"},
			},
		},
		"should associate field with section": {
			input: []*connect.ItemField{
				{
					ID:      "field1",
					Type:    connect.FieldTypeString,
					Value:   "value1",
					Section: &connect.ItemSection{ID: "section1", Label: "Section 1"},
				},
			},
			sectionMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
			expected: []ItemField{
				{
					ID:           "field1",
					Type:         FieldTypeString,
					Value:        "value1",
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
		},
		"should convert all field types": {
			input: []*connect.ItemField{
				{ID: "f1", Type: connect.FieldTypeConcealed, Value: "secret"},
				{ID: "f2", Type: connect.FieldTypeDate, Value: "1609459200"},
				{ID: "f3", Type: connect.FieldTypeEmail, Value: "test@example.com"},
				{ID: "f4", Type: connect.FieldTypeMenu, Value: "option1"},
				{ID: "f5", Type: connect.FieldTypeMonthYear, Value: "202101"},
				{ID: "f6", Type: connect.FieldTypeOTP, Value: "123456"},
				{ID: "f7", Type: connect.FieldTypeString, Value: "text"},
				{ID: "f8", Type: connect.FieldTypeURL, Value: "https://example.com"},
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemField{
				{ID: "f1", Type: FieldTypeConcealed, Value: "secret"},
				{ID: "f2", Type: FieldTypeDate, Value: "2021-01-01"},
				{ID: "f3", Type: FieldTypeEmail, Value: "test@example.com"},
				{ID: "f4", Type: FieldTypeMenu, Value: "option1"},
				{ID: "f5", Type: FieldTypeMonthYear, Value: "202101"},
				{ID: "f6", Type: FieldTypeOTP, Value: "123456"},
				{ID: "f7", Type: FieldTypeString, Value: "text"},
				{ID: "f8", Type: FieldTypeURL, Value: "https://example.com"},
			},
		},
		"should skip nil fields": {
			input: []*connect.ItemField{
				{ID: "field1", Type: connect.FieldTypeString, Value: "value1"},
				nil,
				{ID: "field2", Type: connect.FieldTypeString, Value: "value2"},
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemField{
				{ID: "field1", Type: FieldTypeString, Value: "value1"},
				{ID: "field2", Type: FieldTypeString, Value: "value2"},
			},
		},
		"should handle empty slice": {
			input:      []*connect.ItemField{},
			sectionMap: map[string]ItemSection{},
			expected:   []ItemField{},
		},
		"should error on invalid date timestamp": {
			input: []*connect.ItemField{
				{ID: "date1", Type: connect.FieldTypeDate, Value: "invalid-timestamp"},
			},
			sectionMap:    map[string]ItemSection{},
			expectedError: true,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual, err := fromConnectFields(test.input, test.sectionMap)
			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromConnectFiles(t *testing.T) {
	tests := map[string]struct {
		input      []*connect.File
		sectionMap map[string]ItemSection
		expected   []ItemFile
	}{
		"should convert single file": {
			input: []*connect.File{
				{ID: "file1", Name: "test.txt", Size: 1024, ContentPath: "/path/to/file"},
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemFile{
				{ID: "file1", Name: "test.txt", Size: 1024, ContentPath: "/path/to/file"},
			},
		},
		"should associate file with section": {
			input: []*connect.File{
				{
					ID:          "file1",
					Name:        "test.txt",
					Size:        1024,
					ContentPath: "/path/to/file",
					Section:     &connect.ItemSection{ID: "section1", Label: "Section 1"},
				},
			},
			sectionMap: map[string]ItemSection{
				"section1": {ID: "section1", Label: "Section 1"},
			},
			expected: []ItemFile{
				{
					ID:           "file1",
					Name:         "test.txt",
					Size:         1024,
					ContentPath:  "/path/to/file",
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
		},
		"should skip nil files": {
			input: []*connect.File{
				{ID: "file1", Name: "test1.txt", Size: 100},
				nil,
				{ID: "file2", Name: "test2.txt", Size: 200},
			},
			sectionMap: map[string]ItemSection{},
			expected: []ItemFile{
				{ID: "file1", Name: "test1.txt", Size: 100},
				{ID: "file2", Name: "test2.txt", Size: 200},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := fromConnectFiles(test.input, test.sectionMap)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestToConnectURLs(t *testing.T) {
	tests := map[string]struct {
		input    []ItemURL
		expected []connect.ItemURL
	}{
		"should convert single URL": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
			},
			expected: []connect.ItemURL{
				{URL: "https://example.com", Label: "Example", Primary: true},
			},
		},
		"should convert multiple URLs": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "Primary", Primary: true},
				{URL: "https://example.org", Label: "Secondary", Primary: false},
			},
			expected: []connect.ItemURL{
				{URL: "https://example.com", Label: "Primary", Primary: true},
				{URL: "https://example.org", Label: "Secondary", Primary: false},
			},
		},
		"should handle empty slice": {
			input:    []ItemURL{},
			expected: []connect.ItemURL{},
		},
		"should convert URLs with empty labels": {
			input: []ItemURL{
				{URL: "https://example.com", Label: "", Primary: true},
			},
			expected: []connect.ItemURL{
				{URL: "https://example.com", Label: "", Primary: true},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toConnectURLs(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestToConnectSections(t *testing.T) {
	tests := map[string]struct {
		input    []ItemSection
		expected []*connect.ItemSection
	}{
		"should convert single section": {
			input: []ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
			expected: []*connect.ItemSection{
				{ID: "section1", Label: "Section 1"},
			},
		},
		"should convert multiple sections": {
			input: []ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
			expected: []*connect.ItemSection{
				{ID: "section1", Label: "Section 1"},
				{ID: "section2", Label: "Section 2"},
			},
		},
		"should handle empty slice": {
			input:    []ItemSection{},
			expected: []*connect.ItemSection{},
		},
		"should handle sections with empty labels": {
			input: []ItemSection{
				{ID: "section1", Label: ""},
			},
			expected: []*connect.ItemSection{
				{ID: "section1", Label: ""},
			},
		},
		"should handle sections with empty IDs": {
			input: []ItemSection{
				{ID: "", Label: "Section 1"},
			},
			expected: []*connect.ItemSection{
				{ID: "", Label: "Section 1"},
			},
		},
		"should handle sections with both empty ID and label": {
			input: []ItemSection{
				{ID: "", Label: ""},
			},
			expected: []*connect.ItemSection{
				{ID: "", Label: ""},
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toConnectSections(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Sections mismatch: got %+v, expected %+v", actual, test.expected)
			}
		})
	}
}

func TestToConnectFields(t *testing.T) {
	tests := map[string]struct {
		input         []ItemField
		expected      []*connect.ItemField
		expectedError bool
	}{
		"should convert simple field": {
			input: []ItemField{
				{ID: "field1", Label: "Field 1", Type: FieldTypeString, Value: "value1"},
			},
			expected: []*connect.ItemField{
				{ID: "field1", Label: "Field 1", Type: connect.FieldTypeString, Value: "value1"},
			},
		},
		"should convert date field from YYYY-MM-DD to timestamp": {
			input: []ItemField{
				{ID: "date1", Type: FieldTypeDate, Value: "2021-01-01"},
			},
			expected: []*connect.ItemField{
				{ID: "date1", Type: connect.FieldTypeDate, Value: "1609502460"}, // 2021-01-01 12:01:00 UTC
			},
		},
		"should associate field with section": {
			input: []ItemField{
				{
					ID:           "field1",
					Type:         FieldTypeString,
					Value:        "value1",
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
			expected: []*connect.ItemField{
				{
					ID:    "field1",
					Type:  connect.FieldTypeString,
					Value: "value1",
					Section: &connect.ItemSection{
						ID:    "section1",
						Label: "Section 1",
					},
				},
			},
		},
		"should convert field with generate flag": {
			input: []ItemField{
				{
					ID:       "field1",
					Label:    "Field 1",
					Type:     FieldTypeString,
					Value:    "value1",
					Generate: true,
				},
			},
			expected: []*connect.ItemField{
				{
					ID:       "field1",
					Label:    "Field 1",
					Type:     connect.FieldTypeString,
					Value:    "value1",
					Generate: true,
				},
			},
		},
		"should convert field with recipe and include LETTERS": {
			input: []ItemField{
				{
					ID:   "password",
					Type: FieldTypeConcealed,
					Recipe: &GeneratorRecipe{
						Length:        20,
						CharacterSets: []CharacterSet{CharacterSetDigits, CharacterSetSymbols},
					},
				},
			},
			expected: []*connect.ItemField{
				{
					ID:   "password",
					Type: connect.FieldTypeConcealed,
					Recipe: &connect.GeneratorRecipe{
						Length:        20,
						CharacterSets: []string{"LETTERS", "DIGITS", "SYMBOLS"},
					},
				},
			},
		},
		"should convert all field types": {
			input: []ItemField{
				{ID: "f1", Type: FieldTypeConcealed, Value: "secret"},
				{ID: "f2", Type: FieldTypeDate, Value: "2021-01-01"},
				{ID: "f3", Type: FieldTypeEmail, Value: "test@example.com"},
				{ID: "f4", Type: FieldTypeMenu, Value: "option1"},
				{ID: "f5", Type: FieldTypeMonthYear, Value: "202101"},
				{ID: "f6", Type: FieldTypeOTP, Value: "123456"},
				{ID: "f7", Type: FieldTypeString, Value: "text"},
				{ID: "f8", Type: FieldTypeURL, Value: "https://example.com"},
			},
			expected: []*connect.ItemField{
				{ID: "f1", Type: connect.FieldTypeConcealed, Value: "secret"},
				{ID: "f2", Type: connect.FieldTypeDate, Value: "1609502460"}, // 2021-01-01 12:01:00 UTC
				{ID: "f3", Type: connect.FieldTypeEmail, Value: "test@example.com"},
				{ID: "f4", Type: connect.FieldTypeMenu, Value: "option1"},
				{ID: "f5", Type: connect.FieldTypeMonthYear, Value: "202101"},
				{ID: "f6", Type: connect.FieldTypeOTP, Value: "123456"},
				{ID: "f7", Type: connect.FieldTypeString, Value: "text"},
				{ID: "f8", Type: connect.FieldTypeURL, Value: "https://example.com"},
			},
		},
		"should handle empty slice": {
			input:    []ItemField{},
			expected: []*connect.ItemField{},
		},
		"should error on invalid date string": {
			input: []ItemField{
				{ID: "date1", Type: FieldTypeDate, Value: "invalid-date"},
			},
			expectedError: true,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual, err := toConnectFields(test.input)
			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(actual) != len(test.expected) {
				t.Fatalf("Length mismatch: got %d, expected %d", len(actual), len(test.expected))
			}
			for i := range actual {
				if actual[i].ID != test.expected[i].ID {
					t.Errorf("Fields[%d].ID: got %v, expected %v", i, actual[i].ID, test.expected[i].ID)
				}
				if actual[i].Label != test.expected[i].Label {
					t.Errorf("Fields[%d].Label: got %v, expected %v", i, actual[i].Label, test.expected[i].Label)
				}
				if actual[i].Type != test.expected[i].Type {
					t.Errorf("Fields[%d].Type: got %v, expected %v", i, actual[i].Type, test.expected[i].Type)
				}
				if actual[i].Value != test.expected[i].Value {
					t.Errorf("Fields[%d].Value: got %v, expected %v", i, actual[i].Value, test.expected[i].Value)
				}
				if actual[i].Purpose != test.expected[i].Purpose {
					t.Errorf("Fields[%d].Purpose: got %v, expected %v", i, actual[i].Purpose, test.expected[i].Purpose)
				}
				if actual[i].Generate != test.expected[i].Generate {
					t.Errorf("Fields[%d].Generate: got %v, expected %v", i, actual[i].Generate, test.expected[i].Generate)
				}
				// Compare Section
				if (actual[i].Section == nil) != (test.expected[i].Section == nil) {
					t.Errorf("Fields[%d].Section: got nil=%v, expected nil=%v", i, actual[i].Section == nil, test.expected[i].Section == nil)
				} else if actual[i].Section != nil {
					if actual[i].Section.ID != test.expected[i].Section.ID {
						t.Errorf("Fields[%d].Section.ID: got %v, expected %v", i, actual[i].Section.ID, test.expected[i].Section.ID)
					}
					if actual[i].Section.Label != test.expected[i].Section.Label {
						t.Errorf("Fields[%d].Section.Label: got %v, expected %v", i, actual[i].Section.Label, test.expected[i].Section.Label)
					}
				}
				// Compare Recipe
				if (actual[i].Recipe == nil) != (test.expected[i].Recipe == nil) {
					t.Errorf("Fields[%d].Recipe: got nil=%v, expected nil=%v", i, actual[i].Recipe == nil, test.expected[i].Recipe == nil)
				} else if actual[i].Recipe != nil {
					if actual[i].Recipe.Length != test.expected[i].Recipe.Length {
						t.Errorf("Fields[%d].Recipe.Length: got %v, expected %v", i, actual[i].Recipe.Length, test.expected[i].Recipe.Length)
					}
					if !reflect.DeepEqual(actual[i].Recipe.CharacterSets, test.expected[i].Recipe.CharacterSets) {
						t.Errorf("Fields[%d].Recipe.CharacterSets: got %v, expected %v", i, actual[i].Recipe.CharacterSets, test.expected[i].Recipe.CharacterSets)
					}
				}
			}
		})
	}
}

func TestToConnectFiles(t *testing.T) {
	tests := map[string]struct {
		input    []ItemFile
		expected []*connect.File
	}{
		"should convert single file": {
			input: []ItemFile{
				{ID: "file1", Name: "test.txt", Size: 1024, ContentPath: "/path/to/file"},
			},
			expected: []*connect.File{
				{ID: "file1", Name: "test.txt", Size: 1024, ContentPath: "/path/to/file"},
			},
		},
		"should associate file with section": {
			input: []ItemFile{
				{
					ID:           "file1",
					Name:         "test.txt",
					Size:         1024,
					ContentPath:  "/path/to/file",
					SectionID:    "section1",
					SectionLabel: "Section 1",
				},
			},
			expected: []*connect.File{
				{
					ID:          "file1",
					Name:        "test.txt",
					Size:        1024,
					ContentPath: "/path/to/file",
					Section: &connect.ItemSection{
						ID:    "section1",
						Label: "Section 1",
					},
				},
			},
		},
		"should handle empty slice": {
			input:    []ItemFile{},
			expected: []*connect.File{},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := toConnectFiles(test.input)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
		})
	}
}

func TestFromConnectItemToModel(t *testing.T) {
	tests := map[string]struct {
		input    *connect.Item
		expected *Item
		wantErr  bool
	}{
		"should return error for nil item": {
			input:   nil,
			wantErr: true,
		},
		"should convert basic item": {
			input: &connect.Item{
				ID:       "item1",
				Title:    "Test Item",
				Vault:    connect.ItemVault{ID: "vault1"},
				Category: connect.ItemCategory("LOGIN"),
				Version:  1,
				Tags:     []string{"tag1", "tag2"},
			},
			expected: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				Version:  1,
				Tags:     []string{"tag1", "tag2"},
				URLs:     []ItemURL{},
				Sections: []ItemSection{},
				Fields:   []ItemField{},
				Files:    []ItemFile{},
			},
			wantErr: false,
		},
		"should convert item with sections and fields": {
			input: &connect.Item{
				ID:       "item1",
				Title:    "Test Item",
				Vault:    connect.ItemVault{ID: "vault1"},
				Category: connect.ItemCategory("LOGIN"),
				Sections: []*connect.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []*connect.ItemField{
					{
						ID:      "field1",
						Label:   "Field 1",
						Type:    connect.FieldTypeString,
						Value:   "value1",
						Section: &connect.ItemSection{ID: "section1", Label: "Section 1"},
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
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			item := &Item{}
			err := item.FromConnectItemToModel(test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("FromConnectItemToModel() error = %v, wantErr %v", err, test.wantErr)
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

func TestFromModelItemToConnect(t *testing.T) {
	tests := map[string]struct {
		input    *Item
		expected *connect.Item
		wantErr  bool
	}{
		"should convert basic item": {
			input: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
				Version:  1,
				Tags:     []string{"tag1", "tag2"},
			},
			expected: &connect.Item{
				ID:       "item1",
				Title:    "Test Item",
				Vault:    connect.ItemVault{ID: "vault1"},
				Category: connect.ItemCategory("LOGIN"),
				Version:  1,
				Tags:     []string{"tag1", "tag2"},
				URLs:     []connect.ItemURL{},
				Sections: []*connect.ItemSection{},
				Fields:   []*connect.ItemField{},
				Files:    []*connect.File{},
			},
			wantErr: false,
		},
		"should convert item with sections and fields": {
			input: &Item{
				ID:       "item1",
				Title:    "Test Item",
				VaultID:  "vault1",
				Category: Login,
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
			},
			expected: &connect.Item{
				ID:       "item1",
				Title:    "Test Item",
				Vault:    connect.ItemVault{ID: "vault1"},
				Category: connect.ItemCategory("LOGIN"),
				Sections: []*connect.ItemSection{
					{ID: "section1", Label: "Section 1"},
				},
				Fields: []*connect.ItemField{
					{
						ID:    "field1",
						Label: "Field 1",
						Type:  connect.FieldTypeString,
						Value: "value1",
						Section: &connect.ItemSection{
							ID:    "section1",
							Label: "Section 1",
						},
					},
				},
				Files: []*connect.File{},
			},
			wantErr: false,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual, err := test.input.FromModelItemToConnect()
			if (err != nil) != test.wantErr {
				t.Errorf("FromModelItemToConnect() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			// Compare only the fields we set in our model
			if actual.ID != test.expected.ID {
				t.Errorf("ID: got %v, expected %v", actual.ID, test.expected.ID)
			}
			if actual.Title != test.expected.Title {
				t.Errorf("Title: got %v, expected %v", actual.Title, test.expected.Title)
			}
			if actual.Vault.ID != test.expected.Vault.ID {
				t.Errorf("Vault.ID: got %v, expected %v", actual.Vault.ID, test.expected.Vault.ID)
			}
			if actual.Category != test.expected.Category {
				t.Errorf("Category: got %v, expected %v", actual.Category, test.expected.Category)
			}
			if actual.Version != test.expected.Version {
				t.Errorf("Version: got %v, expected %v", actual.Version, test.expected.Version)
			}
			if !reflect.DeepEqual(actual.Tags, test.expected.Tags) {
				t.Errorf("Tags: got %v, expected %v", actual.Tags, test.expected.Tags)
			}
			if len(actual.URLs) != len(test.expected.URLs) {
				t.Errorf("URLs length: got %d, expected %d", len(actual.URLs), len(test.expected.URLs))
			}
			if len(actual.Sections) != len(test.expected.Sections) {
				t.Errorf("Sections length: got %d, expected %d", len(actual.Sections), len(test.expected.Sections))
			} else {
				for i := range actual.Sections {
					if actual.Sections[i].ID != test.expected.Sections[i].ID {
						t.Errorf("Sections[%d].ID: got %v, expected %v", i, actual.Sections[i].ID, test.expected.Sections[i].ID)
					}
					if actual.Sections[i].Label != test.expected.Sections[i].Label {
						t.Errorf("Sections[%d].Label: got %v, expected %v", i, actual.Sections[i].Label, test.expected.Sections[i].Label)
					}
				}
			}
			if len(actual.Fields) != len(test.expected.Fields) {
				t.Errorf("Fields length: got %d, expected %d", len(actual.Fields), len(test.expected.Fields))
			} else {
				for i := range actual.Fields {
					if actual.Fields[i].ID != test.expected.Fields[i].ID {
						t.Errorf("Fields[%d].ID: got %v, expected %v", i, actual.Fields[i].ID, test.expected.Fields[i].ID)
					}
					if actual.Fields[i].Label != test.expected.Fields[i].Label {
						t.Errorf("Fields[%d].Label: got %v, expected %v", i, actual.Fields[i].Label, test.expected.Fields[i].Label)
					}
					if actual.Fields[i].Type != test.expected.Fields[i].Type {
						t.Errorf("Fields[%d].Type: got %v, expected %v", i, actual.Fields[i].Type, test.expected.Fields[i].Type)
					}
					if actual.Fields[i].Value != test.expected.Fields[i].Value {
						t.Errorf("Fields[%d].Value: got %v, expected %v", i, actual.Fields[i].Value, test.expected.Fields[i].Value)
					}
					if (actual.Fields[i].Section == nil) != (test.expected.Fields[i].Section == nil) {
						t.Errorf("Fields[%d].Section: got nil=%v, expected nil=%v", i, actual.Fields[i].Section == nil, test.expected.Fields[i].Section == nil)
					} else if actual.Fields[i].Section != nil {
						if actual.Fields[i].Section.ID != test.expected.Fields[i].Section.ID {
							t.Errorf("Fields[%d].Section.ID: got %v, expected %v", i, actual.Fields[i].Section.ID, test.expected.Fields[i].Section.ID)
						}
						if actual.Fields[i].Section.Label != test.expected.Fields[i].Section.Label {
							t.Errorf("Fields[%d].Section.Label: got %v, expected %v", i, actual.Fields[i].Section.Label, test.expected.Fields[i].Section.Label)
						}
					}
				}
			}
			if len(actual.Files) != len(test.expected.Files) {
				t.Errorf("Files length: got %d, expected %d", len(actual.Files), len(test.expected.Files))
			}
		})
	}
}
