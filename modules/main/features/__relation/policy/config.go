package relation

import (
	"context"

	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
)

type RefListConfig struct {
	Permissions []string
	RefDTO      any
	CachePrefix string
}

type SearchHandler func(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[any], error)

type RefSearchConfig struct {
	Permissions []string
	RefDTO      any
	CachePrefix string
}

type Config struct {
	// Schema
	MainTable string // ví dụ: "materials"
	RefTable  string // ví dụ: "suppliers"

	// Runtime: cách lấy ID từ entity + input
	// entity: struct chứa ID chính (vd: material)
	// input:  struct chứa danh sách IDs (vd: SupplierIDs)
	GetMainID func(entity any) (int, error)
	GetIDs    func(input any) ([]int, error)
	SetResult func(output any, resIDs []int, resAsStr *string, res []string) error

	GetRefList   *RefListConfig
	GetRefSearch *RefSearchConfig
}
