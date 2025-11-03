package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/khiemnd777/andy_api/modules/auth_facebook/config"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/repository"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/service/fblimited"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
	tokenApi "github.com/khiemnd777/andy_api/shared/modules/token"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthFacebookService struct {
	repo              *repository.AuthFacebookRepository
	deps              *module.ModuleDeps[config.ModuleConfig]
	fbLimitedVerifier *fblimited.Verifier
	secret            string
	refreshTTL        time.Duration
	accessTTL         time.Duration
}

func NewAuthFacebookService(repo *repository.AuthFacebookRepository, deps *module.ModuleDeps[config.ModuleConfig], secret string) *AuthFacebookService {
	return &AuthFacebookService{
		repo:   repo,
		deps:   deps,
		secret: secret,
		fbLimitedVerifier: func() *fblimited.Verifier {
			v, err := fblimited.NewVerifier(context.Background(), deps)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create FB limited verifier: %v", err))
				return nil
			}
			return v
		}(),
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute,
	}
}

type facebookGraphResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

func (s *AuthFacebookService) FBLimitedVerify(ctx context.Context, jwt string) (*auth.AuthTokenPair, error) {
	claims, err := s.fbLimitedVerifier.Verify(ctx, jwt)

	if err != nil {
		return nil, err
	}

	fbUserID := claims.Subject
	email := claims.Email
	name := claims.Name
	dummyAvatar := utils.GetDummyAvatarURL(name)

	user, isNew, err := s.repo.FindOrCreateUserByFacebook(ctx, &repository.FacebookUserInfo{
		Email:   email,
		Name:    name,
		Picture: dummyAvatar,
		Sub:     fbUserID,
	})

	if err != nil {
		return nil, err
	}

	tokens, err := tokenApi.GenerateTokens(ctx, user.ID)

	if err != nil {
		return nil, err
	}

	if isNew {
		// Assign default role
		pubsub.PublishAsync("role:default", utils.AssignDefaultRole{
			UserID:  user.ID,
			RoleIDs: []int{1},
		})
	}

	return tokens, nil
}

func (s *AuthFacebookService) LoginWithFacebook(ctx context.Context, accessToken string) (*auth.AuthTokenPair, error) {
	info, err := s.getFacebookUserInfo(accessToken)

	if err != nil {
		return nil, err
	}

	user, isNew, err := s.repo.FindOrCreateUserByFacebook(ctx, &repository.FacebookUserInfo{
		Email:   info.Email,
		Name:    info.Name,
		Picture: info.Picture.Data.URL,
		Sub:     info.ID,
	})

	if err != nil {
		return nil, err
	}

	tokens, err := tokenApi.GenerateTokens(ctx, user.ID)

	if err != nil {
		return nil, err
	}

	if isNew {
		// Assign default role
		pubsub.PublishAsync("role:default", utils.AssignDefaultRole{
			UserID:  user.ID,
			RoleIDs: []int{1},
		})
	}

	return tokens, nil
}

func (s *AuthFacebookService) getFacebookUserInfo(accessToken string) (*facebookGraphResponse, error) {
	url := fmt.Sprintf("%s&access_token=%s", s.deps.Config.AuthFacebook.UserInfoURL, accessToken)
	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid facebook token")
	}
	defer res.Body.Close()

	var data facebookGraphResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
