package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"identity/config"
	"identity/domain"
	"time"
)

type TokenService struct {
	config *config.JWTConfig
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Roles     []string  `json:"roles"`
	jwt.RegisteredClaims
}

func NewTokenService(cfg *config.JWTConfig) *TokenService {
	return &TokenService{config: cfg}
}

func (t *TokenService) GenerateAccessToken(user *domain.User) (string, error) {
	var roles []string
	for _, ur := range user.UserRoles {
		roles = append(roles, ur.Role.Name)
	}

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.config.Issuer,
			Audience:  []string{t.config.Audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.config.ExpiryMinutes)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.config.SecretKey))
}

func (t *TokenService) GenerateRefreshToken(userID uuid.UUID, ip string) (*domain.RefreshToken, error) {
	// Generate random token
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	token := base64.URLEncoding.EncodeToString(bytes)

	return &domain.RefreshToken{
		UserID:      userID,
		Token:       token,
		ExpiresAt:   time.Now().Add(t.config.RefreshExpiryDays),
		CreatedByIP: ip,
		IsRevoked:   false,
	}, nil
}

func (t *TokenService) ValidateRefreshToken(token *domain.RefreshToken) error {
	if token == nil {
		return errors.New("invalid refresh token")
	}
	if !token.IsActive() {
		return errors.New("refresh token is expired or revoked")
	}
	return nil
}
