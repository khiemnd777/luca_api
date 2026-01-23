package engine

import (
	"fmt"
	"time"

	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

func (e *Engine) matchConditions(
	promo *generated.PromotionCode,
	orderCtx OrderContext,
	now time.Time,
) ([]string, error) {

	var applied []string

	for _, cond := range promo.Edges.Conditions {
		switch cond.ConditionType {

		case promotionmodel.PromotionConditionOrderIsRemake:
			if !orderCtx.IsRemake {
				return nil, PromotionApplyError{Reason: ReasonConditionOrderIsRemakeNotMet}
			}

		case promotionmodel.PromotionConditionRemakeCountLTE:
			value, err := parseIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeCount > value {
				return nil, PromotionApplyError{Reason: ReasonConditionRemakeCountLTENotMet}
			}

		case promotionmodel.PromotionConditionRemakeWithinDays:
			value, err := parseIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.OriginalTime.IsZero() {
				return nil, PromotionApplyError{Reason: ReasonConditionRemakeWithinDaysNotMet}
			}
			if int(now.Sub(orderCtx.OriginalTime).Hours()/24) > value {
				return nil, PromotionApplyError{Reason: ReasonConditionRemakeWithinDaysNotMet}
			}

		case promotionmodel.PromotionConditionRemakeReason:
			values, err := parseStringList(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeReason == "" || !containsString(values, orderCtx.RemakeReason) {
				return nil, PromotionApplyError{Reason: ReasonConditionRemakeReasonNotMet}
			}

		default:
			return nil, fmt.Errorf("unsupported condition type: %s", cond.ConditionType)
		}

		applied = append(applied, string(cond.ConditionType))
	}

	return applied, nil
}
