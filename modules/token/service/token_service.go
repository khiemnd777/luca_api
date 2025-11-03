package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/token/repository"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type TokenService struct {
	repo       *repository.TokenRepository
	secret     string
	refreshTTL time.Duration
	accessTTL  time.Duration
}

func NewTokenService(repo *repository.TokenRepository, secret string) *TokenService {
	return &TokenService{
		repo:       repo,
		secret:     secret,
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute,
	}
}

func (s *TokenService) GenerateTokens(ctx context.Context, id int) (*auth.AuthTokenPair, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	access, err := utils.GenerateJWTToken(s.secret, utils.JWTTokenPayload{
		UserID: user.ID,
		Email:  user.Email,
		Exp:    time.Now().Add(s.accessTTL),
	})
	if err != nil {
		return nil, err
	}

	refresh, err := utils.GenerateJWTToken(s.secret, utils.JWTTokenPayload{
		UserID: user.ID,
		Email:  user.Email,
		Exp:    time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return nil, err
	}

	err = s.repo.CreateRefreshToken(ctx, user.ID, refresh, time.Now().Add(s.refreshTTL))
	if err != nil {
		return nil, err
	}

	return &auth.AuthTokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *TokenService) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthTokenPair, error) {
	found, valid, userID, email, err := s.repo.IsRefreshTokenValid(ctx, refreshToken)

	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if found && !valid {
		return nil, errors.New("invalid refresh token")
	}

	newAccessToken, tokenErr := utils.GenerateJWTToken(s.secret, utils.JWTTokenPayload{
		UserID: userID,
		Email:  email,
		Exp:    time.Now().Add(s.accessTTL),
	})
	if tokenErr != nil {
		return nil, tokenErr
	}

	newRefreshToken, tokenErr := utils.GenerateJWTToken(s.secret, utils.JWTTokenPayload{
		UserID: userID,
		Email:  email,
		Exp:    time.Now().Add(s.refreshTTL),
	})
	if tokenErr != nil {
		return nil, tokenErr
	}

	err = s.repo.CreateRefreshToken(ctx, userID, newRefreshToken, time.Now().Add(s.refreshTTL))
	if err != nil {
		return nil, err
	}

	return &auth.AuthTokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *TokenService) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *TokenService) CleanupExpiredRefreshTokens(ctx context.Context) error {
	return s.repo.DeleteExpiredRefreshTokens(ctx)
}
