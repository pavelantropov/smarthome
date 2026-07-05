package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"identity/internal/config"
	"identity/internal/domain"
	"identity/internal/dto"
	"identity/internal/repository"
	"log/slog"
	"time"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	refreshRepo  *repository.RefreshTokenRepository
	tokenService *TokenService
	passwordSvc  *PasswordService
	config       *config.Config
	logger       *slog.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	refreshRepo *repository.RefreshTokenRepository,
	tokenService *TokenService,
	passwordSvc *PasswordService,
	config *config.Config,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		refreshRepo:  refreshRepo,
		tokenService: tokenService,
		passwordSvc:  passwordSvc,
		config:       config,
		logger:       logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest, ip string) (*dto.AuthResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.IsEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	exists, err = s.userRepo.IsUsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, errors.New("username already taken")
	}

	// Hash password
	passwordHash, err := s.passwordSvc.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PhoneNumber:  req.PhoneNumber,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default role (User)
	// For simplicity, we'll assign role ID 1 (User)
	// In production, you should get role by name
	// roleID := uuid.MustParse("...")

	s.logger.Info("User registered successfully",
		"user_id", user.ID,
		"username", user.Username,
		"email", user.Email,
	)

	return s.generateAuthResponse(ctx, user, ip)
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip string) (*dto.AuthResponse, error) {
	// Get user by username or email
	user, err := s.userRepo.GetByUsernameOrEmail(ctx, req.UsernameOrEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid username/email or password")
	}

	// Verify password
	valid, err := s.passwordSvc.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, errors.New("invalid username/email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Update last login
	user.LastLoginAt = nil
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Warn("failed to update last login", "error", err)
	}

	s.logger.Info("User logged in successfully",
		"user_id", user.ID,
		"username", user.Username,
	)

	return s.generateAuthResponse(ctx, user, ip)
}

func (s *AuthService) RefreshToken(ctx context.Context, token string, ip string) (*dto.AuthResponse, error) {
	// Get refresh token
	refreshToken, err := s.refreshRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Validate refresh token
	if err := s.tokenService.ValidateRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Revoke old refresh token
	refreshToken.IsRevoked = true
	now := time.Now()
	refreshToken.RevokedAt = &now
	refreshToken.RevokedByIP = ip
	refreshToken.Reason = "Refreshed"

	if err := s.refreshRepo.Update(ctx, refreshToken); err != nil {
		s.logger.Warn("failed to revoke old refresh token", "error", err)
	}

	s.logger.Info("Token refreshed successfully",
		"user_id", user.ID,
		"username", user.Username,
	)

	return s.generateAuthResponse(ctx, user, ip)
}

func (s *AuthService) Logout(ctx context.Context, token string, ip string) error {
	refreshToken, err := s.refreshRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return nil // Token already invalid
	}

	refreshToken.IsRevoked = true
	now := time.Now()
	refreshToken.RevokedAt = &now
	refreshToken.RevokedByIP = ip
	refreshToken.Reason = "Logout"

	return s.refreshRepo.Update(ctx, refreshToken)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *domain.User, ip string) (*dto.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Get roles
	var roles []string
	for _, ur := range user.UserRoles {
		roles = append(roles, ur.Role.Name)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresAt:    time.Now().Add(s.config.JWT.ExpiryMinutes),
		User: dto.UserDTO{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			PhoneNumber: user.PhoneNumber,
			Roles:       roles,
		},
	}, nil
}
