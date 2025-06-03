package auth

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"server/config"
	"server/pkg/failure"
	"time"
)

type TokenCache interface {
	Store(ctx context.Context, token string, ttl time.Duration) error
	Check(ctx context.Context, token string) (bool, error)
	Del(ctx context.Context, token string) error
}

type AuthService struct {
	tokens TokenCache

	adminUsername string
	adminPassword string
	tokenTTL      time.Duration
}

func NewAuthService(tokenCache TokenCache, cfg *config.AuthConfig) *AuthService {
	service := &AuthService{
		tokens:        tokenCache,
		adminUsername: cfg.Username,
		adminPassword: cfg.Password,
		tokenTTL:      time.Hour * 24 * time.Duration(cfg.TokenTTLDays),
	}

	return service
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	if login == s.adminUsername && password == s.adminPassword {
		uid := uuid.New().String()

		if err := s.tokens.Store(ctx, uid, s.tokenTTL); err != nil {
			return "", err
		}

		return uid, nil
	}

	return "", failure.NewUnauthorized("invalid login or password")
}

func (s *AuthService) Logout(ctx context.Context, uid string) error {
	return s.tokens.Del(ctx, uid)
}

func (s *AuthService) CheckSession(ctx context.Context, uid string) (bool, error) {
	ok, err := s.tokens.Check(ctx, uid)
	if err != nil {
		return ok, err
	}

	return ok, nil
}
