package fblimited

import (
	"context"
	"errors"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/config"
	"github.com/khiemnd777/andy_api/shared/module"
)

type Claims struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

type Verifier struct {
	jwks   *keyfunc.JWKS
	appID  string
	issuer string
}

func NewVerifier(ctx context.Context, deps *module.ModuleDeps[config.ModuleConfig]) (*Verifier, error) {
	clientID := deps.Config.AuthFacebook.ClientID
	const jwksURL = "https://www.facebook.com/.well-known/oauth/openid/jwks/"
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Verifier{
		jwks:   jwks,
		appID:  clientID,
		issuer: "https://www.facebook.com",
	}, nil
}

func (v *Verifier) Verify(ctx context.Context, tokenStr string) (*Claims, error) {
	var c Claims
	tok, err := jwt.ParseWithClaims(tokenStr, &c, v.jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	// iss
	if c.Issuer != v.issuer && c.Issuer != "https://facebook.com" {
		return nil, errors.New("invalid issuer")
	}
	// aud
	if !c.VerifyAudience(v.appID, true) {
		return nil, errors.New("invalid audience")
	}
	// exp/nbf/iAT
	if err := c.RegisteredClaims.Valid(); err != nil {
		return nil, err
	}
	return &c, nil
}
