package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/modules/realtime"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	searchutils "github.com/khiemnd777/andy_api/shared/search"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderService interface {
	Create(ctx context.Context, deptID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	Update(ctx context.Context, deptID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	UpdateStatus(ctx context.Context, orderItemProcessID int64, status string) (*model.OrderItemDTO, error)
	GetByID(ctx context.Context, id int64) (*model.OrderDTO, error)
	GetByOrderIDAndOrderItemID(ctx context.Context, orderID, orderItemID int64) (*model.OrderDTO, error)
	PrepareForRemakeByOrderID(ctx context.Context, orderID int64) (*model.OrderDTO, error)
	GetAllOrderProducts(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error)
	GetAllOrderMaterials(ctx context.Context, orderID int64) ([]*model.OrderItemMaterialDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderDTO], error)
	InProgressList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.InProcessOrderDTO], error)
	NewestList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.NewestOrderDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error)
	Delete(ctx context.Context, id int64) error
	SyncPrice(ctx context.Context, orderID int64) (float64, error)
}

type orderService struct {
	repo  repository.OrderRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewOrderService(repo repository.OrderRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderService {
	return &orderService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kOrderByID(id int64) string {
	return fmt.Sprintf("order:id:%d", id)
}

func kOrderByIDAll(id int64) string {
	return fmt.Sprintf("order:id:%d:*", id)
}

func kOrderAll() []string {
	return []string{
		kOrderListAll(),
		kOrderSearchAll(),
		"order:assigned:*",
		"order:item:material:loaner:*",
		"order:list:inprogress:*",
		"order:list:newest:*",
	}
}

func kOrderListAll() string {
	return "order:list:*"
}

func kOrderSearchAll() string {
	return "order:search:*"
}

func kOrderList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("order:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kOrderInProgressList(q table.TableQuery) string {
	return fmt.Sprintf("order:list:inprogress:l%d:p%d", q.Limit, q.Page)
}

func kOrderNewestList(q table.TableQuery) string {
	return fmt.Sprintf("order:list:newest:l%d:p%d", q.Limit, q.Page)
}

func kOrderSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("order:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *orderService) Create(ctx context.Context, deptID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kOrderByID(dto.ID), kOrderByIDAll(dto.ID))
	}
	cache.InvalidateKeys(kOrderAll()...)

	s.upsertSearch(ctx, deptID, dto)

	realtime.BroadcastAll("order:newest", nil)

	return dto, nil
}

func (s *orderService) Update(ctx context.Context, deptID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kOrderByID(dto.ID), kOrderByIDAll(dto.ID))
	}
	cache.InvalidateKeys(kOrderAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, orderItemProcessID int64, status string) (*model.OrderItemDTO, error) {
	out, err := s.repo.UpdateStatus(ctx, orderItemProcessID, status)
	if err != nil {
		return nil, err
	}

	if out != nil {
		cache.InvalidateKeys(
			kOrderByID(out.OrderID),
			kOrderByIDAll(out.OrderID),
		)
	}
	cache.InvalidateKeys(kOrderAll()...)

	return out, nil
}

func (s *orderService) upsertSearch(ctx context.Context, deptID int, dto *model.OrderDTO) {
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "order", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "order",
		EntityID:   int64(dto.ID),
		Title:      *dto.Code,
		Subtitle:   nil,
		Keywords:   &kwPtr,
		Content:    nil,
		Attributes: map[string]any{},
		OrgID:      utils.Ptr(int64(deptID)),
		OwnerID:    nil,
	})
}

func (s *orderService) unlinkSearch(id int64) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "order",
		EntityID:   id,
	})
}

func (s *orderService) GetByID(ctx context.Context, id int64) (*model.OrderDTO, error) {
	return s.repo.GetByID(ctx, id)
	// return cache.Get(kOrderByID(id), cache.TTLMedium, func() (*model.OrderDTO, error) {
	// 	return s.repo.GetByID(ctx, id)
	// })
}

func (s *orderService) GetByOrderIDAndOrderItemID(ctx context.Context, orderID, orderItemID int64) (*model.OrderDTO, error) {
	return s.repo.GetByOrderIDAndOrderItemID(ctx, orderID, orderItemID)
	// return cache.Get(kOrderByOrderIDAndOrderItemID(orderID, orderItemID), cache.TTLMedium, func() (*model.OrderDTO, error) {
	// 	return s.repo.GetByOrderIDAndOrderItemID(ctx, orderID, orderItemID)
	// })
}

func (s *orderService) PrepareForRemakeByOrderID(ctx context.Context, orderID int64) (*model.OrderDTO, error) {
	return s.repo.PrepareForRemakeByOrderID(ctx, orderID)
}

func (s *orderService) GetAllOrderProducts(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error) {
	return s.repo.GetAllOrderProducts(ctx, orderID)
}

func (s *orderService) GetAllOrderMaterials(ctx context.Context, orderID int64) ([]*model.OrderItemMaterialDTO, error) {
	return s.repo.GetAllOrderMaterials(ctx, orderID)
}

func (s *orderService) SyncPrice(ctx context.Context, orderID int64) (float64, error) {
	return s.repo.SyncPrice(ctx, orderID)
}

func (s *orderService) NewestList(ctx context.Context, q table.TableQuery) (table.TableListResult[model.NewestOrderDTO], error) {
	type boxed = table.TableListResult[model.NewestOrderDTO]

	query := q
	query.OrderBy = utils.Ptr("created_at")
	query.Direction = "desc"
	key := kOrderNewestList(query)

	ptr, err := cache.Get(key, cache.TTLShort, func() (*boxed, error) {
		list, err := s.repo.NewestList(ctx, query)
		if err != nil {
			return nil, err
		}
		return &list, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *orderService) InProgressList(ctx context.Context, q table.TableQuery) (table.TableListResult[model.InProcessOrderDTO], error) {
	type boxed = table.TableListResult[model.InProcessOrderDTO]

	query := q
	query.OrderBy = utils.Ptr("delivery_date")
	query.Direction = "desc"
	key := kOrderInProgressList(query)

	ptr, err := cache.Get(key, cache.TTLShort, func() (*boxed, error) {
		list, err := s.repo.InProgressList(ctx, query)
		if err != nil {
			return nil, err
		}
		return &list, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	now := time.Now()
	items := make([]*model.InProcessOrderDTO, 0, len(ptr.Items))
	for _, item := range ptr.Items {
		if item == nil {
			items = append(items, nil)
			continue
		}
		itemCopy := *item
		itemCopy.Now = &now
		items = append(items, &itemCopy)
	}
	res := *ptr
	res.Items = items
	return res, nil
}

func (s *orderService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.OrderDTO], error) {
	type boxed = table.TableListResult[model.OrderDTO]
	key := kOrderList(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, q)
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

func (s *orderService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kOrderAll()...)
	cache.InvalidateKeys(kOrderByID(id))

	s.unlinkSearch(id)
	return nil
}

func (s *orderService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error) {
	type boxed = dbutils.SearchResult[model.OrderDTO]
	key := kOrderSearch(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, q)
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
