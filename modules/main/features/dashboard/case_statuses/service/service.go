package service

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_statuses/repository"
	"github.com/khiemnd777/andy_api/shared/module"
)

var CaseStatusTargets = map[string]int{
	"received":    24,
	"in_progress": 18,
	"qc":          12,
	"issue":       8,
	"rework":      10,
}

type caseStatusMetaItem struct {
	Label  string
	Color  string
	Helper string
}

var caseStatusMeta = map[string]caseStatusMetaItem{
	"received": {
		Label:  "received",
		Color:  "#64748B",
		Helper: "New cases received today",
	},
	"in_progress": {
		Label:  "in_progress",
		Color:  "#2563EB",
		Helper: "Cases currently in progress",
	},
	"qc": {
		Label:  "qc",
		Color:  "#0EA5E9",
		Helper: "Cases in quality control",
	},
	"issue": {
		Label:  "issue",
		Color:  "#DC2626",
		Helper: "Cases blocked by issues",
	},
	"rework": {
		Label:  "rework",
		Color:  "#D97706",
		Helper: "Cases requiring rework",
	},
}

type CaseStatusesService interface {
	CaseStatuses(
		ctx context.Context,
		deptID int,
	) ([]*model.CaseStatusItem, error)
}

type caseStatusesService struct {
	repo repository.CaseStatusesRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseStatusesService(
	repo repository.CaseStatusesRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseStatusesService {
	return &caseStatusesService{repo: repo, deps: deps}
}

func BuildCaseStatuses(
	raw []*model.CaseStatusCount,
) []*model.CaseStatusItem {
	countMap := make(map[string]int, len(raw))
	for _, r := range raw {
		countMap[r.Status] = r.Count
	}

	result := make([]*model.CaseStatusItem, 0, len(caseStatusMeta))
	for status, meta := range caseStatusMeta {
		result = append(result, &model.CaseStatusItem{
			Status: status,
			Label:  meta.Label,
			Count:  countMap[status],
			Target: CaseStatusTargets[status],
			Color:  meta.Color,
			Helper: meta.Helper,
		})
	}

	return result
}

func (s *caseStatusesService) CaseStatuses(
	ctx context.Context,
	deptID int,
) ([]*model.CaseStatusItem, error) {
	raw, err := s.repo.CountByStatus(ctx, deptID)
	if err != nil {
		return nil, err
	}

	return BuildCaseStatuses(raw), nil
}
