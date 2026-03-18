package repository

import (
	"strings"
	"testing"

	moduleModel "github.com/khiemnd777/andy_api/modules/observability/model"
	sharedconfig "github.com/khiemnd777/andy_api/shared/config"
)

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

func TestParseLogEntry_FallbackFromStreamLabels(t *testing.T) {
	raw := `{"ts":"2026-03-17T10:00:00Z","level":"warn","message":"fallback"}`
	stream := map[string]string{
		"service":       "luca_api",
		"module":        "gateway",
		"env":           "local",
		"request_id":    "req-stream",
		"user_id":       "99",
		"department_id": "3",
	}

	entry := parseLogEntry("1742205600000000000", raw, stream)

	if entry.Service != "luca_api" {
		t.Fatalf("expected service from stream label, got %q", entry.Service)
	}
	if entry.Module != "gateway" {
		t.Fatalf("expected module from stream label, got %q", entry.Module)
	}
	if entry.Environment != "local" {
		t.Fatalf("expected env from stream label, got %q", entry.Environment)
	}
	if entry.RequestID != "req-stream" {
		t.Fatalf("expected request_id from stream label, got %q", entry.RequestID)
	}
	if entry.UserID != 99 {
		t.Fatalf("expected user_id from stream label, got %d", entry.UserID)
	}
	if entry.DepartmentID != 3 {
		t.Fatalf("expected department_id from stream label, got %d", entry.DepartmentID)
	}
}

func TestBuildStreamQuery_IncludesJSONErrorFilterAndOptionalFilters(t *testing.T) {
	userID := 42
	deptID := 7
	repo := &observabilityRepository{
		cfg: sharedconfig.ObservabilityConfig{
			Loki: sharedconfig.ObservabilityLokiConfig{
				StreamSelector: `{app="luca_api"}`,
			},
		},
	}

	query := repo.buildStreamQuery(moduleQueryForTest(userID, deptID))

	expected := []string{
		`{app="luca_api"}`,
		`| json`,
		`| __error__=""`,
		`| level=~"warn|error"`,
		`| module="auth"`,
		`| service="luca_api"`,
		`| env="local"`,
		`| request_id="req-1"`,
		`| user_id="42"`,
		`| department_id="7"`,
	}
	for _, part := range expected {
		if !strings.Contains(query, part) {
			t.Fatalf("expected query to contain %q, got %q", part, query)
		}
	}
}

func TestNormalizeLimit_UsesConfigMaxQueryLimit(t *testing.T) {
	repo := &observabilityRepository{
		cfg: sharedconfig.ObservabilityConfig{
			Loki: sharedconfig.ObservabilityLokiConfig{
				MaxQueryLimit: 100,
			},
		},
	}

	if got := repo.normalizeLimit(200); got != 100 {
		t.Fatalf("expected clamped limit 100, got %d", got)
	}
	if got := repo.normalizeLimit(0); got != 50 {
		t.Fatalf("expected default limit 50, got %d", got)
	}
}

func moduleQueryForTest(userID, deptID int) moduleModel.ListLogsQuery {
	return moduleModel.ListLogsQuery{
		Levels:    []string{"warn", "error"},
		Module:    "auth",
		Service:   "luca_api",
		Env:       "local",
		RequestID: "req-1",
		UserID:    &userID,
		DeptID:    &deptID,
	}
}
