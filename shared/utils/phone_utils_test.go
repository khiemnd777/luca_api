package utils

import "testing"

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{input: Ptr("+84974322365"), expected: "+84974322365"},
		{input: Ptr("0974322365"), expected: "+84974322365"},
		{input: Ptr("84974322365"), expected: "+84974322365"},
		{input: Ptr("+84 974 322 365"), expected: "+84974322365"},
		{input: Ptr("0974.322.365"), expected: "+84974322365"},
		{input: nil, expected: ""},
	}

	for _, tt := range tests {
		actual := NormalizePhone(tt.input)
		if actual != tt.expected {
			t.Errorf("NormalizePhone(%v) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestIsPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"+84974322365", true},
		{"0974322365", true},
		{"84974322365", true},
		{"+84 974 322 365", true},
		{"0974.322.365", true},
		{"abc123", false},
		{"0123", false},
		{"+1-800-555-1234", false},
	}

	for _, tt := range tests {
		if got := IsPhone(tt.input); got != tt.expected {
			t.Errorf("IsPhone(%q) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}
