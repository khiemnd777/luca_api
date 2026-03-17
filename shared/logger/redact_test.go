package logger

import "testing"

func TestRedactFields(t *testing.T) {
	input := map[string]any{
		"token":    "abc",
		"password": "123",
		"nested": map[string]any{
			"authorization": "Bearer secret",
			"safe":          "ok",
		},
	}

	output := redactFields(input, []string{"token", "password", "authorization"})

	if output["token"] != "***redacted***" {
		t.Fatalf("expected token to be redacted")
	}
	if output["password"] != "***redacted***" {
		t.Fatalf("expected password to be redacted")
	}

	nested, ok := output["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map to remain a map")
	}
	if nested["authorization"] != "***redacted***" {
		t.Fatalf("expected nested authorization to be redacted")
	}
	if nested["safe"] != "ok" {
		t.Fatalf("expected safe field to remain unchanged")
	}
}
