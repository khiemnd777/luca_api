package product_api

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type DeductQuantityItem struct {
	ProductID    int `json:"product_id"`
	RequestedQty int `json:"requested_qty"`
}

type BatchDeductRequest struct {
	Items []DeductQuantityItem `json:"items"`
}

func BatchDeductQuantity(ctx context.Context, accessToken string, items []DeductQuantityItem) error {
	payload := BatchDeductRequest{Items: items}

	err := app.GetHttpClient().CallPost(ctx,
		"product",
		"/api/product/batch-deduct-quantity",
		accessToken,
		"",
		payload, nil,
	)

	if err != nil {
		logger.Warn("❌ Batch deduct quantity failed: %v", err)
		return err
	}

	return nil
}
