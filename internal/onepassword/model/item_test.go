package model

import (
	"testing"

	sdk "github.com/1password/onepassword-sdk-go"
)

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
