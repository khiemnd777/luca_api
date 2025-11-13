package service

import (
	"context"

	"github.com/khiemnd777/andy_api/modules/search/config"
	"github.com/khiemnd777/andy_api/modules/search/model"
	"github.com/khiemnd777/andy_api/modules/search/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	sharedmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
)

type SearchService interface {
	Upsert(ctx context.Context, d sharedmodel.Doc) error
	Delete(ctx context.Context, entityType string, entityID int64) error
	Search(ctx context.Context, opt model.Options) ([]model.Row, error)
}

type searchService struct {
	repo repository.SearchRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSearchService(repo repository.SearchRepository, deps *module.ModuleDeps[config.ModuleConfig]) SearchService {
	return &searchService{repo: repo, deps: deps}
}

func (r *searchService) Upsert(ctx context.Context, d sharedmodel.Doc) error {
	return r.repo.Upsert(ctx, d)
}

func (r *searchService) Delete(ctx context.Context, entityType string, entityID int64) error {
	return r.repo.Delete(ctx, entityType, entityID)
}

func (r *searchService) Search(ctx context.Context, opt model.Options) ([]model.Row, error) {
	return r.repo.Search(ctx, opt)
}
