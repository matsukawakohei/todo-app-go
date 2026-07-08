package usecase

import (
	"strings"
	"testing"
)

func TestLabelCreateValidate(t *testing.T) {
	tests := []struct {
		testName string
		hasError bool
		name     string
		expected string
	}{
		{
			testName: "nameEmpty",
			hasError: true,
			name:     "",
			expected: "name is required",
		},
		{
			testName: "nameMinChars",
			hasError: false,
			name:     "あ",
		},
		{
			testName: "nameMaxChars",
			hasError: false,
			name:     strings.Repeat("あ", 50),
		},
		{
			testName: "nameOverChars",
			hasError: true,
			name:     strings.Repeat("あ", 51),
			expected: "limited max 50 char",
		},
	}

	v := NewLabelValidator()
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			input := LabelCreateInput{Name: tc.name}

			err := v.CreateValidate(t.Context(), input)
			if tc.hasError {
				if err == nil {
					t.Fatalf("expect err is not nil; but got nil")
				}
				if !strings.Contains(err.Error(), tc.expected) {
					t.Errorf("expect err contains %v; but got %v", tc.expected, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expect err is nul; but got %v", err)
				}
			}
		})
	}
}
