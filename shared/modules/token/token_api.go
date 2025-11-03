package tokenApi

import (
	"context"
	"net/http"
	"time"

	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func GenerateTokens(ctx context.Context, userId int) (*auth.AuthTokenPair, error) {
	generateTokenData := map[string]any{
		"userID": userId,
	}
	var result auth.AuthTokenPair
	err := app.GetHttpClient().CallRequest(
		ctx,
		http.MethodPost,
		"token",
		"/api/token/generate",
		"",
		utils.GetInternalToken(),
		generateTokenData,
		&result,
		app.RetryOptions{
			MaxAttempts: 3,
			Delay:       200 * time.Millisecond,
		},
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func RefreshTokens(ctx context.Context, refreshToken string) (*auth.AuthTokenPair, error) {
	refreshTokenData := map[string]any{
		"refreshToken": refreshToken,
	}
	var result auth.AuthTokenPair
	err := app.GetHttpClient().CallRequest(
		ctx,
		http.MethodPost,
		"token",
		"/api/token/refresh",
		"",
		utils.GetInternalToken(),
		refreshTokenData,
		&result,
		app.RetryOptions{
			MaxAttempts: 3,
			Delay:       200 * time.Millisecond,
		},
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	deleteRefreshTokenData := map[string]any{
		"refreshToken": refreshToken,
	}
	return app.GetHttpClient().CallRequest(
		ctx,
		http.MethodDelete,
		"token",
		"/api/token/delete",
		"",
		utils.GetInternalToken(),
		deleteRefreshTokenData,
		nil,
		app.RetryOptions{
			MaxAttempts: 3,
			Delay:       200 * time.Millisecond,
		},
	)
}
