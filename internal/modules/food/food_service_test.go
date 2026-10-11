package food

import (
	"testing"
)

func TestDetermineNutritionBasis(t *testing.T) {
	tests := []struct {
		name          string
		explicitBasis string
		servingSize   string
		expected      string
	}{
		{
			name:          "Explicit PER_100G",
			explicitBasis: "PER_100G",
			servingSize:   "100 g",
			expected:      "PER_100G",
		},
		{
			name:          "Serving size contains 100 g without explicit basis",
			explicitBasis: "",
			servingSize:   "100 g",
			expected:      "PER_100G",
		},
		{
			name:          "Serving size contains 100 gram",
			explicitBasis: "",
			servingSize:   "100 gram",
			expected:      "PER_100G",
		},
		{
			name:          "Serving size is 20.30 g (portion/sajian)",
			explicitBasis: "",
			servingSize:   "20.30 g",
			expected:      "PER_SERVING",
		},
		{
			name:          "Serving size is 1 porsi",
			explicitBasis: "",
			servingSize:   "1 porsi",
			expected:      "PER_SERVING",
		},
		{
			name:          "Serving size is 1 botol / kemasan",
			explicitBasis: "",
			servingSize:   "1 botol (250 ml)",
			expected:      "PER_PACKAGE",
		},
		{
			name:          "Serving size contains kemasan",
			explicitBasis: "",
			servingSize:   "Per Kemasan",
			expected:      "PER_PACKAGE",
		},
		{
			name:          "Serving size is 0 g (invalid/unverified)",
			explicitBasis: "",
			servingSize:   "0 g",
			expected:      "UNVERIFIED",
		},
		{
			name:          "Serving size is empty (unverified)",
			explicitBasis: "",
			servingSize:   "",
			expected:      "UNVERIFIED",
		},
		{
			name:          "Serving size is 0",
			explicitBasis: "",
			servingSize:   "0",
			expected:      "UNVERIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineNutritionBasis(tt.explicitBasis, tt.servingSize)
			if got != tt.expected {
				t.Errorf("determineNutritionBasis(%q, %q) = %q; want %q", tt.explicitBasis, tt.servingSize, got, tt.expected)
			}
		})
	}
}
