package service

import (
	"fmt"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	policy "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

func init() {
	policy.Register("material", policy.Config{
		MainTable: "materials",
		RefTable:  "suppliers",

		GetMainID: func(entity any) (int, error) {
			e, ok := entity.(*generated.Material)
			if !ok || e == nil {
				return 0, fmt.Errorf("relation material: entity is not *generated.Material")
			}
			return e.ID, nil
		},

		GetIDs: func(input any) ([]int, error) {
			in, ok := input.(model.MaterialDTO)
			if !ok {
				return nil, fmt.Errorf("relation material: input is not *model.MaterialDTO")
			}
			return in.SupplierIDs, nil
		},

		SetResult: func(output any, res []string) error {
			out, ok := output.(*model.MaterialDTO)
			if !ok || out == nil {
				return fmt.Errorf("relation material: output is not *model.MaterialDTO")
			}

			out.SupplierNames = res

			return nil
		},
	})
}
