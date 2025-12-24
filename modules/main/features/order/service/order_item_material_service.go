package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderItemMaterialService interface {
	GetLoanerMaterials(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemMaterialDTO], error)
}

type orderItemMaterialService struct {
	repo repository.OrderItemMaterialRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderItemMaterialService(
	repo repository.OrderItemMaterialRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
) OrderItemMaterialService {
	return &orderItemMaterialService{repo: repo, deps: deps}
}

func kOrderItemLoanerMaterialList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("order:item:material:loaner:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func (s *orderItemMaterialService) GetLoanerMaterials(
	ctx context.Context,
	query table.TableQuery,
) (table.TableListResult[model.OrderItemMaterialDTO], error) {
	type boxed = table.TableListResult[model.OrderItemMaterialDTO]
	key := kOrderItemLoanerMaterialList(query)

	ptr, err := cache.Get(key, cache.TTLShort, func() (*boxed, error) {
		res, e := s.repo.GetLoanerMaterials(ctx, query)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}
