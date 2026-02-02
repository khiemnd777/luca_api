package model

import "time"

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
