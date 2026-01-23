package engine

import "errors"

type PromotionApplyError struct {
	Reason string
}

func (e PromotionApplyError) Error() string {
	return e.Reason
}

func IsPromotionApplyError(err error) (string, bool) {
	var perr PromotionApplyError
	if errors.As(err, &perr) {
		return perr.Reason, true
	}
	return "", false
}
