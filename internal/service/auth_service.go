package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/EdOoO21/openapi-and-crud/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users      *repository.UserRepository
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(u *repository.UserRepository) *AuthService {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev_secret_change_me"
	}
	accessMin := 20
	if s := os.Getenv("ACCESS_TOKEN_TTL_MIN"); s != "" {
		if v, _ := strconv.Atoi(s); v > 0 {
			accessMin = v
		}
	}
	refreshDays := 7
	if s := os.Getenv("REFRESH_TOKEN_TTL_DAYS"); s != "" {
		if v, _ := strconv.Atoi(s); v > 0 {
			refreshDays = v
		}
	}
	return &AuthService{
		users:      u,
		secret:     sec,
		accessTTL:  time.Duration(accessMin) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (*repository.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.users.CreateUser(ctx, email, string(hash), role)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, *repository.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	access, err := s.createAccessToken(u)
	if err != nil {
		return nil, nil, err
	}
	refreshID := uuid.New()
	expires := time.Now().Add(s.refreshTTL)
	if err := s.users.CreateRefreshToken(ctx, refreshID, u.ID, "", expires); err != nil {
		return nil, nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refreshID.String()}, u, nil
}

func (s *AuthService) createAccessToken(u *repository.User) (string, error) {
	now := time.Now()
	claims := &tokenClaims{
		UserID: u.ID.String(),
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.secret))
}

func (s *AuthService) Refresh(ctx context.Context, refreshStr string) (*TokenPair, error) {
	id, err := uuid.Parse(refreshStr)
	if err != nil {
		return nil, err
	}
	rec, err := s.users.GetRefreshToken(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.ExpiresAt.Before(time.Now()) {
		_ = s.users.DeleteRefreshToken(ctx, id)
		return nil, errors.New("refresh expired")
	}
	u, err := s.users.GetByID(ctx, rec.UserID)
	if err != nil {
		return nil, err
	}
	access, err := s.createAccessToken(u)
	if err != nil {
		return nil, err
	}
	newID := uuid.New()
	expires := time.Now().Add(s.refreshTTL)
	if err := s.users.DeleteRefreshToken(ctx, id); err != nil {
		return nil, err
	}
	if err := s.users.CreateRefreshToken(ctx, newID, u.ID, "", expires); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: newID.String()}, nil
}
