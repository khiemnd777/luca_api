package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	authErrors "github.com/khiemnd777/andy_api/modules/auth/model/error"
	"github.com/khiemnd777/andy_api/modules/auth/repository"
	"github.com/khiemnd777/andy_api/shared/auth"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	tokenApi "github.com/khiemnd777/andy_api/shared/modules/token"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthService struct {
	repo       *repository.AuthRepository
	secret     string
	refreshTTL time.Duration
	accessTTL  time.Duration
}

func NewAuthService(repo *repository.AuthRepository, secret string) *AuthService {
	return &AuthService{
		repo:       repo,
		secret:     secret,
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute,
	}
}

func (s *AuthService) CreateNewUser(ctx context.Context, phoneOrEmail, password, name string) (*generated.User, error) {
	var phone *string
	var email *string

	switch {
	case utils.IsEmail(phoneOrEmail):
		email = &phoneOrEmail
		if exists, _ := s.repo.CheckEmailExists(ctx, *email); exists {
			return nil, authErrors.ErrPhoneOrEmailExists
		}
	case utils.IsPhone(phoneOrEmail):
		normalizedPhone := utils.NormalizePhone(&phoneOrEmail)
		phone = &normalizedPhone

		if exists, _ := s.repo.CheckPhoneExists(ctx, *phone); exists {
			return nil, authErrors.ErrPhoneOrEmailExists
		}
	default:
		return nil, authErrors.ErrPhoneOrEmailExists
	}

	dummyAvatar := utils.GetDummyAvatarURL(name)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	refCode := uuid.NewString()
	qrCode := utils.GenerateQRCodeStringForUser(refCode)

	user, err := s.repo.CreateNewUser(ctx, phone, email, string(hashedPassword), name, &dummyAvatar, &refCode, &qrCode)

	if err != nil {
		return nil, err
	}

	// Assign default role
	pubsub.PublishAsync("role:default", utils.AssignDefaultRole{
		UserID:  user.ID,
		RoleIDs: []int{1},
	})

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, phoneOrEmail, password string) (*auth.AuthTokenPair, error) {
	var (
		tokens *auth.AuthTokenPair
		err    error
	)
	switch {
	case utils.IsEmail(phoneOrEmail):
		_, tokens, err = s.LoginWithEmail(ctx, phoneOrEmail, password)
	case utils.IsPhone(phoneOrEmail):
		_, tokens, err = s.LoginWithPhone(ctx, phoneOrEmail, password)
	default:
		return nil, authErrors.ErrInvalidCredentials
	}

	return tokens, err
}

func (s *AuthService) LoginWithPhone(ctx context.Context, phone, password string) (*generated.User, *auth.AuthTokenPair, error) {
	user, err := s.repo.GetUserByPhone(ctx, phone)

	if err != nil {
		return nil, nil, authErrors.ErrInvalidCredentials
	}

	return s.loginUser(ctx, user, password)
}

func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (*generated.User, *auth.AuthTokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)

	if err != nil {
		return user, nil, authErrors.ErrInvalidCredentials
	}

	return s.loginUser(ctx, user, password)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthTokenPair, error) {
	tokens, err := tokenApi.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return tokenApi.DeleteRefreshToken(ctx, refreshToken)
}

func (s *AuthService) loginUser(ctx context.Context, user *generated.User, password string) (*generated.User, *auth.AuthTokenPair, error) {
	if user == nil {
		return nil, nil, authErrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, authErrors.ErrInvalidCredentials
	}

	tokens, err := tokenApi.GenerateTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}
