package service

import (
	"fmt"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/validator"
)

func validatePromotionInput(input *model.CreatePromotionInput) error {
	for i, s := range input.Scopes {
		if err := validator.ValidateScopeValue(s.ScopeType, s.ScopeValue); err != nil {
			return fmt.Errorf("scopes[%d]: %w", i, err)
		}
	}

	for i, c := range input.Conditions {
		if err := validator.ValidateConditionValue(c.ConditionType, c.ConditionValue); err != nil {
			return fmt.Errorf("conditions[%d]: %w", i, err)
		}
	}

	return nil
}
