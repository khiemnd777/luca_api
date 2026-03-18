package service

import (
	"context"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/modules/observability/model"
	"github.com/khiemnd777/andy_api/modules/observability/repository"
)

type ObservabilityService interface {
	ListLogs(ctx context.Context, query model.ListLogsQuery) ([]model.LogEntry, error)
	GetSummary(ctx context.Context, query model.ListLogsQuery) (model.Summary, error)
}

type observabilityService struct {
	repo repository.ObservabilityRepository
}

func NewObservabilityService(repo repository.ObservabilityRepository) ObservabilityService {
	return &observabilityService{repo: repo}
}

func (s *observabilityService) ListLogs(ctx context.Context, query model.ListLogsQuery) ([]model.LogEntry, error) {
	query = normalizeQuery(query)
	return s.repo.ListLogs(ctx, query)
}

func (s *observabilityService) GetSummary(ctx context.Context, query model.ListLogsQuery) (model.Summary, error) {
	query = normalizeQuery(query)
	return s.repo.GetSummary(ctx, query)
}

func normalizeQuery(query model.ListLogsQuery) model.ListLogsQuery {
	query.Direction = strings.ToUpper(strings.TrimSpace(query.Direction))
	if query.Direction == "" {
		query.Direction = "BACKWARD"
	}

	if query.Limit <= 0 {
		query.Limit = 50
	}

	if query.To.IsZero() {
		query.To = time.Now()
	}
	if query.From.IsZero() {
		query.From = query.To.Add(-1 * time.Hour)
	}
	if len(query.Levels) == 0 {
		query.Levels = []string{"warn", "error"}
	}

	for i, level := range query.Levels {
		query.Levels[i] = strings.ToLower(strings.TrimSpace(level))
	}
	query.Module = strings.TrimSpace(query.Module)
	query.Service = strings.TrimSpace(query.Service)
	query.Env = strings.TrimSpace(query.Env)
	query.RequestID = strings.TrimSpace(query.RequestID)
	query.Keyword = strings.TrimSpace(query.Keyword)
	if query.UserID != nil && *query.UserID < 0 {
		query.UserID = nil
	}
	if query.DeptID != nil && *query.DeptID < 0 {
		query.DeptID = nil
	}

	return query
}
