package handler

import (
	"github.com/gin-gonic/gin"
	"identity/internal/dto"
	"identity/internal/service"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ip := c.ClientIP()
	response, err := h.authService.Register(c.Request.Context(), &req, ip)
	if err != nil {
		h.logger.Error("Registration failed", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ip := c.ClientIP()
	response, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		h.logger.Error("Login failed", "error", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ip := c.ClientIP()
	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, ip)
	if err != nil {
		h.logger.Error("Token refresh failed", "error", err)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ip := c.ClientIP()
	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken, ip); err != nil {
		h.logger.Error("Logout failed", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
