package engine

import (
	"encoding/json"
	"time"
)

type Promotion struct {
	ID                int
	Code              string
	IsActive          bool
	StartAt           time.Time
	EndAt             time.Time
	DiscountType      string
	DiscountValue     int
	MinOrderValue     *int
	MaxDiscountAmount *int
	Scopes            []PromotionScope
	Conditions        []PromotionCondition
}

type PromotionScope struct {
	Type  string
	Value json.RawMessage
}

type PromotionCondition struct {
	Type  string
	Value json.RawMessage
}
