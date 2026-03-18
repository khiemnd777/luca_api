package model

import "time"

type ListLogsQuery struct {
	Levels    []string
	Module    string
	Service   string
	Env       string
	RequestID string
	UserID    *int
	DeptID    *int
	Keyword   string
	Direction string
	From      time.Time
	To        time.Time
	Limit     int
}

type LogEntry struct {
	Timestamp    time.Time      `json:"ts"`
	Level        string         `json:"level"`
	Message      string         `json:"message"`
	Service      string         `json:"service,omitempty"`
	Module       string         `json:"module,omitempty"`
	Environment  string         `json:"env,omitempty"`
	RequestID    string         `json:"request_id,omitempty"`
	UserID       int            `json:"user_id,omitempty"`
	DepartmentID int            `json:"department_id,omitempty"`
	Method       string         `json:"method,omitempty"`
	Path         string         `json:"path,omitempty"`
	Source       string         `json:"source,omitempty"`
	Error        string         `json:"error,omitempty"`
	Stacktrace   string         `json:"stacktrace,omitempty"`
	Fields       map[string]any `json:"fields,omitempty"`
	Raw          string         `json:"raw,omitempty"`
}

type Summary struct {
	From   time.Time      `json:"from"`
	To     time.Time      `json:"to"`
	Counts map[string]int `json:"counts"`
}
