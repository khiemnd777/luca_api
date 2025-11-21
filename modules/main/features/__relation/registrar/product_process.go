package registrar

import (
	"fmt"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func init() {
	logger.Debug("[RELATION] Register products - processes")
	policy.Register("product",
		policy.Config{
			MainTable: "products",
			RefTable:  "processes",

			GetMainID: func(entity any) (int, error) {
				e, ok := entity.(*generated.Product)
				if !ok || e == nil {
					return 0, fmt.Errorf("relation product: entity is not *generated.Product")
				}
				return e.ID, nil
			},

			GetIDs: func(input any) ([]int, error) {
				in, ok := input.(model.ProductDTO)
				if !ok {
					return nil, fmt.Errorf("relation product: input is not *model.ProductDTO")
				}
				return in.ProcessIDs, nil
			},

			SetResult: func(output any, ids []int, resStr *string, res []string) error {
				out, ok := output.(*model.ProductDTO)
				if !ok || out == nil {
					return fmt.Errorf("relation product: output is not *model.ProductDTO")
				}

				out.ProcessNames = resStr

				return nil
			},

			GetRefList: &policy.RefListConfig{
				Permissions: []string{"process.view"},
				RefDTO:      model.ProcessDTO{},
				CachePrefix: "process:list",
			},
		},
	)
}
