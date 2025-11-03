package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/auditlog/config"
	"github.com/khiemnd777/andy_api/modules/auditlog/ent/generated"
	"github.com/khiemnd777/andy_api/modules/auditlog/model"
	"github.com/khiemnd777/andy_api/modules/auditlog/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuditLogService struct {
	repo *repository.AuditLogRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuditLogService(repo *repository.AuditLogRepository, deps *module.ModuleDeps[config.ModuleConfig]) *AuditLogService {
	return &AuditLogService{repo: repo, deps: deps}
}

func (s *AuditLogService) Log(ctx context.Context, userID int, action, module string, targetID int, data map[string]any) error {
	return cache.UpdateAndInvalidate(fmt.Sprintf("module:%s:target:%d:list:first-page", module, targetID), func() error {
		log := &generated.AuditLog{
			UserID:   userID,
			Action:   action,
			Module:   module,
			TargetID: &targetID,
			Data:     data,
		}
		return s.repo.Create(ctx, log)
	})
}

func (s *AuditLogService) ListByTargetPaginated(ctx context.Context, module string, targetID, limit, page int) ([]*model.AuditLogModel, bool, error) {
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	if page == 1 {
		key := fmt.Sprintf("module:%s:target:%d:list:first-page", module, targetID)
		return cache.GetListWithHasMore(key, cache.TTLShort, func() ([]*model.AuditLogModel, bool, error) {
			return s.repo.ListByTargetPaginated(ctx, module, targetID, limit, offset)
		})
	}

	return s.repo.ListByTargetPaginated(ctx, module, targetID, limit, offset)
}
