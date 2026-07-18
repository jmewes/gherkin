package utils

import "testing"

func TestConvertStringToKebabCase(t *testing.T) {

	testCases := []struct {
		name     string
		s        string
		expected string
	}{

		{
			name:     "feature name",
			s:        "User authentication",
			expected: "user-authentication",
		},
		{
			name:     "with special characters",
			s:        "Tax rate of 5.5%",
			expected: "tax-rate-of-5-5",
		},
		{
			name:     "single word",
			s:        "Bar",
			expected: "bar",
		},
		{
			name:     "upper camel case",
			s:        "JustATest",
			expected: "just-a-test",
		},
		{
			name:     "lower camel case",
			s:        "justATest",
			expected: "just-a-test",
		},
		{
			name:     "TBD",
			s:        "resource_Name_test",
			expected: "resource-name-test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given a string <s>
			s := tc.s

			// When the string is converted to kebab-case
			result := ToKebabCase(s)

			// Then the resulting string is all lower-case
			// And has everything that is not a character or digit with a dash,
			// And leading and trailing digits are removed.
			if result != tc.expected {
				t.Errorf("expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}
