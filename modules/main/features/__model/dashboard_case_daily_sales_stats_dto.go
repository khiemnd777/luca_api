package model

import (
	"time"
)

type SalesSummary struct {
	TotalRevenue    float64  `json:"total_revenue,omitempty"`
	OrderItemsCount int      `json:"order_items_count,omitempty"`
	PrevRevenue     float64  `json:"prev_revenue,omitempty"`
	GrowthPercent   *float64 `json:"growth_percent,omitempty"`
}

type SalesDailyItem struct {
	Date    time.Time `json:"date,omitempty"`
	Revenue float64   `json:"revenue,omitempty"`
}

type Range string

const (
	RangeToday Range = "today"
	Range7d    Range = "7d"
	Range30d   Range = "30d"
)

type SalesReportResponse struct {
	KPIs *SalesSummary     `json:"kpis,omitempty"`
	Line []*SalesDailyItem `json:"line,omitempty"`
}

type SalesDailyUpsert struct {
	StatAt       time.Time
	DepartmentID int
}
