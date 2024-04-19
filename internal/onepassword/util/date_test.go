package util

import "testing"

func TestIsValidDateFormat(t *testing.T) {
	tests := map[string]struct {
		date     string
		expected bool
	}{
		"should return \"false\" if secondsStr is empty string": {
			date:     "",
			expected: false,
		},
		"should return \"false\" if secondsStr invalid secondsStr string": {
			date:     "20-20-20",
			expected: false,
		},
		"should return \"false\" if secondsStr in DD-MM-YYYY format": {
			date:     "20-12-2024",
			expected: false,
		},
		"should return \"false\" if secondsStr in MM-DD-YYYY format": {
			date:     "12-20-2024",
			expected: false,
		},
		"should return \"false\" if secondsStr in YYYY-MM-DDHHmmss format": {
			date:     "12-20-2024T12:20:00",
			expected: false,
		},
		"should return \"true\" if secondsStr has valid format": {
			date:     "2024-12-31",
			expected: true,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			actual := IsValidDateFormat(test.date)
			if actual != test.expected {
				t.Errorf("Expected %t for %s secondsStr, got %t", test.expected, test.date, actual)
			}
		})
	}
}

func TestSecondsToYYYYMMDD(t *testing.T) {
	tests := map[string]struct {
		secondsStr string
		expected   string
		expectErr  bool
	}{
		"should return error if secondsStr is empty string": {
			secondsStr: "",
			expectErr:  true,
		},
		"should return error if secondsStr contains not number chars": {
			secondsStr: "1234abcd",
			expectErr:  true,
		},
		"should return \"YYYY-MM-DD\" string": {
			secondsStr: "1696914000",
			expected:   "2023-10-10",
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			date, err := SecondsToYYYYMMDD(test.secondsStr)
			if err != nil && !test.expectErr {
				t.Errorf("Expected no error but got one for secondsStr: %s", test.secondsStr)
			}
			if date != test.expected {
				t.Errorf("Expected %s for %s secondsStr, got %s", test.expected, test.secondsStr, date)
			}
		})
	}
}
