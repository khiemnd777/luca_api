package validator

import (
	"encoding/json"
	"errors"
	"fmt"

	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
)

// ----- Condition

func ValidateConditionValue(
	condType promotionmodel.PromotionConditionType,
	raw json.RawMessage,
) error {

	validator, ok := conditionValidators[condType]
	if !ok {
		return fmt.Errorf("unsupported condition type: %s", condType)
	}

	if err := promotionmodel.ValidateConditionInput(condType); err != nil {
		return err
	}

	if err := validator(raw); err != nil {
		return fmt.Errorf("invalid condition_value for %s: %w", condType, err)
	}

	return nil
}

type conditionValidator func(raw json.RawMessage) error

var conditionValidators = map[promotionmodel.PromotionConditionType]conditionValidator{

	promotionmodel.PromotionConditionOrderIsRemake: func(raw json.RawMessage) error {
		if len(raw) != 0 && string(raw) != "null" {
			return errors.New("ORDER_IS_REMAKE must not have condition_value")
		}
		return nil
	},

	promotionmodel.PromotionConditionRemakeCountLTE: func(raw json.RawMessage) error {
		v, err := parseInt(raw)
		if err != nil {
			return err
		}
		if v < 0 {
			return errors.New("REMAKE_COUNT_LTE must be >= 0")
		}
		return nil
	},

	promotionmodel.PromotionConditionRemakeWithinDays: func(raw json.RawMessage) error {
		v, err := parseInt(raw)
		if err != nil {
			return err
		}
		if v <= 0 {
			return errors.New("REMAKE_WITHIN_DAYS must be > 0")
		}
		return nil
	},

	promotionmodel.PromotionConditionRemakeReason: func(raw json.RawMessage) error {
		_, err := parseStringList(raw)
		return err
	},
}

// ----- Scope

func ValidateScopeValue(
	scopeType promotionmodel.PromotionScopeType,
	raw json.RawMessage,
) error {

	validator, ok := scopeValidators[scopeType]
	if !ok {
		return fmt.Errorf("unsupported scope type: %s", scopeType)
	}

	if err := promotionmodel.ValidateScopeInput(scopeType); err != nil {
		return err
	}

	if err := validator(raw); err != nil {
		return fmt.Errorf("invalid scope_value for %s: %w", scopeType, err)
	}

	return nil
}

type scopeValidator func(raw json.RawMessage) error

var scopeValidators = map[promotionmodel.PromotionScopeType]scopeValidator{

	promotionmodel.PromotionScopeAll: func(raw json.RawMessage) error {
		if len(raw) != 0 && string(raw) != "null" {
			return errors.New("ALL scope must not have scope_value")
		}
		return nil
	},

	promotionmodel.PromotionScopeUser: func(raw json.RawMessage) error {
		_, err := parseIntListAllowNull(raw)
		return err
	},

	promotionmodel.PromotionScopeSeller: func(raw json.RawMessage) error {
		_, err := parseIntListAllowNull(raw)
		return err
	},

	promotionmodel.PromotionScopeProduct: func(raw json.RawMessage) error {
		_, err := parseIntListAllowNull(raw)
		return err
	},

	promotionmodel.PromotionScopeCategory: func(raw json.RawMessage) error {
		_, err := parseIntListAllowNull(raw)
		return err
	},
}

// ----- Helper
func parseInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, errors.New("value is required")
	}

	var v int
	if err := json.Unmarshal(raw, &v); err == nil {
		return v, nil
	}

	return 0, errors.New("expected integer")
}

func parseStringList(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, errors.New("value is required")
	}

	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		if len(list) == 0 {
			return nil, errors.New("string list cannot be empty")
		}
		return list, nil
	}

	return nil, errors.New("expected string array")
}

func parseIntListAllowNull(raw json.RawMessage) ([]int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var list []int
	if err := json.Unmarshal(raw, &list); err == nil {
		if len(list) == 0 {
			return nil, errors.New("int list cannot be empty")
		}
		return list, nil
	}

	return nil, errors.New("expected int array")
}
