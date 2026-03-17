package repository

import "testing"

func TestParseLogEntry(t *testing.T) {
	raw := `{"ts":"2026-03-17T10:00:00Z","level":"error","message":"boom","service":"luca_api","module":"auth","request_id":"req-1","user_id":42,"department_id":7,"source":"foo.go:10","error":"db down","stacktrace":"trace","path":"/api/x","method":"POST","extra":"value"}`

	entry := parseLogEntry("1742205600000000000", raw, map[string]string{})

	if entry.Level != "error" {
		t.Fatalf("expected level=error, got %q", entry.Level)
	}
	if entry.Message != "boom" {
		t.Fatalf("expected message boom, got %q", entry.Message)
	}
	if entry.RequestID != "req-1" {
		t.Fatalf("expected request_id req-1, got %q", entry.RequestID)
	}
	if entry.Fields["extra"] != "value" {
		t.Fatalf("expected extra field to be preserved")
	}
}
