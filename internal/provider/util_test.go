package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestSetTimeValue(t *testing.T) {
	tests := map[string]struct {
		input    time.Time
		expected basetypes.StringValue
	}{
		"should format a UTC timestamp as RFC 3339": {
			input:    time.Date(2024, 3, 1, 10, 30, 0, 0, time.UTC),
			expected: types.StringValue("2024-03-01T10:30:00Z"),
		},
		"should convert a timestamp with an offset to UTC": {
			input:    time.Date(2024, 3, 1, 10, 30, 0, 0, time.FixedZone("EST", -5*60*60)),
			expected: types.StringValue("2024-03-01T15:30:00Z"),
		},
		"should return null for a zero timestamp": {
			input:    time.Time{},
			expected: types.StringNull(),
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := setTimeValue(test.input)
			if !actual.Equal(test.expected) {
				t.Errorf("setTimeValue() = %v, expected %v", actual, test.expected)
			}
		})
	}
}
