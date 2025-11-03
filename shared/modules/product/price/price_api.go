package priceApi

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type TotalProductPriceRequest struct {
	ProductID       int     `json:"product_id"`
	Quantity        int     `json:"quantity"`
	PreOrderPercent float64 `json:"pre_order_percent"`
}

type totalPriceResponse struct {
	TotalPrice float64 `json:"total_price"`
}

type totalPriceRequest struct {
	Items []TotalProductPriceRequest `json:"items"`
}

func GetTotalPrice(ctx context.Context, accessToken string, items []TotalProductPriceRequest) (float64, error) {
	payload := totalPriceRequest{Items: items}
	var resp totalPriceResponse
	err := app.GetHttpClient().CallPost(ctx, "product", "/api/product/price/total", accessToken, "", payload, &resp)

	if err != nil {
		logger.Warn("Pricing service failed or returned nil. Fallback to totalPrice = 0")
		return 0.0, err
	}

	return resp.TotalPrice, nil
}

func GetRawTotalPrice(ctx context.Context, accessToken string, items []TotalProductPriceRequest) (float64, error) {
	payload := totalPriceRequest{Items: items}
	var resp totalPriceResponse
	err := app.GetHttpClient().CallPost(ctx, "product", "/api/product/price/raw-total", accessToken, "", payload, &resp)

	if err != nil {
		logger.Warn("Pricing service failed or returned nil. Fallback to totalPrice = 0")
		return 0.0, err
	}

	return resp.TotalPrice, nil
}

type productPriceResponse struct {
	ProductID int     `json:"product_id"`
	Price     float64 `json:"price"`
}

func GetPrice(ctx context.Context, accessToken string, productID int) (float64, error) {
	var resp productPriceResponse
	url := fmt.Sprintf("/api/product/%d/price", productID)

	err := app.GetHttpClient().CallGet(ctx, "product", url, accessToken, "", &resp)
	if err != nil {
		logger.Warn("Failed to fetch product price", "product_id", productID, "err", err)
		return 0.0, err
	}

	return resp.Price, nil
}
