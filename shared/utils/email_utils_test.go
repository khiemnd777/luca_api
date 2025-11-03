package utils

import "testing"

func TestIsEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"user@example.com", true},
		{"User@Example.com", true},
		{"user.name+tag@gmail.com", true},
		{"invalid@", false},
		{"@invalid.com", false},
		{"justtext", false},
		{"1234567890", false},
		{"user@localhost", false},
		{"user@domain.co", true},
		{"user@domain.c", false},
	}

	for _, tt := range tests {
		result := IsEmail(tt.input)
		if result != tt.expected {
			t.Errorf("IsEmail(%q) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}
