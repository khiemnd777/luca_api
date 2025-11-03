package repository

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/auditlog/config"
	"github.com/khiemnd777/andy_api/modules/auditlog/ent/generated"
	"github.com/khiemnd777/andy_api/modules/auditlog/ent/generated/auditlog"
	"github.com/khiemnd777/andy_api/modules/auditlog/model"
	sharedGenerated "github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuditLogRepository struct {
	db       *generated.Client
	sharedDB *sharedGenerated.Client
	deps     *module.ModuleDeps[config.ModuleConfig]
}

func NewAuditLogRepository(db *generated.Client, sharedDB *sharedGenerated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AuditLogRepository {
	return &AuditLogRepository{db: db, sharedDB: sharedDB, deps: deps}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *generated.AuditLog) error {
	return r.db.AuditLog.Create().
		SetUserID(log.UserID).
		SetAction(log.Action).
		SetModule(log.Module).
		SetNillableTargetID(log.TargetID).
		SetData(log.Data).
		Exec(ctx)
}

func (r *AuditLogRepository) ListByTargetPaginated(ctx context.Context, module string, targetID, limit, offset int) ([]*model.AuditLogModel, bool, error) {
	query, err := r.db.AuditLog.
		Query().
		Where(auditlog.ModuleEQ(module), auditlog.TargetIDEQ(targetID)).
		Order(generated.Desc(auditlog.FieldCreatedAt)).
		Offset(offset).
		Limit(limit + 1).
		All(ctx)

	if err != nil {
		return nil, false, err
	}

	hasMore := len(query) > limit
	if hasMore {
		query = query[:limit]
	}

	var result []*model.AuditLogModel

	for _, n := range query {
		result = append(result, &model.AuditLogModel{
			ID:        n.ID,
			UserID:    n.UserID,
			Action:    n.Action,
			Module:    n.Module,
			TargetID:  n.TargetID,
			Data:      n.Data,
			CreatedAt: n.CreatedAt,
		})

		targetUser, err := r.sharedDB.User.
			Query().
			Where(user.ID(n.UserID)).
			Only(ctx)

		if err == nil {
			result[len(result)-1].User = targetUser
		}
	}

	logger.Debug(fmt.Sprintf("ListByTargetPaginated: module=%s, targetID=%d, limit=%d, offset=%d, hasMore=%v, itemsCount=%d",
		module, targetID, limit, offset, hasMore, len(result)))

	return result, hasMore, nil
}
