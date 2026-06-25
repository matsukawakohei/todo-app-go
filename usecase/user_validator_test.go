package usecase

import (
	"context"
	"strings"
	"testing"
)

func TestSignUpValidate(t *testing.T) {
	tests := []struct {
		testName string
		hasError bool
		expected string
		name string
		email string
		password string
	}{
		{
			testName: "nameEmpty",
			hasError: true,
			expected: "Name: name is required.",
			name: "",
			email: "test@example.com",
			password: "Password1234",
		},
		{
			testName: "nameMinChars",
			hasError: false,
			expected: "",
			name: "a",
			email: "test@example.com",
			password: "Password1234",
		},
		{
			testName: "nameMaxChars",
			hasError: false,
			expected: "",
			name: strings.Repeat("あ", 50),
			email: "test@example.com",
			password: "Password1234",
		},
		{
			testName: "nameOverChars",
			hasError: true,
			expected: "Name: limited max 50 char.",
			name: strings.Repeat("あ", 51),
			email: "test@example.com",
			password: "Password1234",
		},
		{
			testName: "emailRequred",
			hasError: true,
			expected: "Email: email is required.",
			name: "テスト",
			email: "",
			password: "Password1234",
		},
		{
			testName: "emailMaxChars",
			hasError: false,
			expected: "",
			name: "テスト",
			email: strings.Repeat("a", 38) + "@example.com",
			password: "Password1234",
		},
		{
			testName: "emailMaxChars",
			hasError: true,
			expected: "Email: limited max 50 char.",
			name: "テスト",
			email: strings.Repeat("a", 39) + "@example.com",
			password: "Password1234",
		},
		{
			testName: "emailMaxChars",
			hasError: true,
			expected: "Email: is not valid email format.",
			name: "テスト",
			email: "test",
			password: "Password1234",
		},
		{
			testName: "passwordRequired",
			hasError: true,
			expected: "Password: password is required.",
			name: "テスト",
			email: "test@example.com",
			password: "",
		},
		{
			testName: "passwordBelowMinChars",
			hasError: true,
			expected: "Password: limited min 6 max 30 char.",
			name: "テスト",
			email: "test@example.com",
			password: "Abc12",
		},
		{
			testName: "passwordMinChars",
			hasError: false,
			expected: "",
			name: "テスト",
			email: "test@example.com",
			password: "Abcd1E",
		},
		{
			testName: "passwordMaxChars",
			hasError: false,
			expected: "",
			name: "テスト",
			email: "test@example.com",
			password: "A" + strings.Repeat("a", 28) + "1",
		},
		{
			testName: "passwordOverChars",
			hasError: true,
			expected: "Password: limited min 6 max 30 char.",
			name: "テスト",
			email: "test@example.com",
			password: "A" + strings.Repeat("a", 29) + "1",
		},
		{
			testName: "passwordNoLowerCase",
			hasError: true,
			expected: "Password: required lower case.",
			name: "テスト",
			email: "test@example.com",
			password: "PASSWORD123",
		},
		{
			testName: "passwordNoUpperCase",
			hasError: true,
			expected: "Password: required upper case.",
			name: "テスト",
			email: "test@example.com",
			password: "password123",
		},
		{
			testName: "passwordNoNumber",
			hasError: true,
			expected: "Password: required number.",
			name: "テスト",
			email: "test@example.com",
			password: "Password",
		},
		{
			testName: "passwordInvalidFormat",
			hasError: true,
			expected: "Password: invalid format.",
			name: "テスト",
			email: "test@example.com",
			password: "Password1!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			input := SignUpInput{
				Name: tc.name,
				Email: tc.email,
				Password: tc.password,
			}
			v := NewUserValidator()
			
			err := v.SignUpValidate(context.Background(), input)
			if tc.hasError {
				if err.Error() != tc.expected {
					t.Errorf("expect err is %v; but got %v", tc.expected, err)
				}
			} else {
				if err != nil {
					t.Errorf("expect err is nil; but got %v", err)
				}
			}
			
		})
	}
}