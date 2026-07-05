package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"identity/domain"
	"log/slog"
	"os"

	"identity/config"
	"identity/handler"
	"identity/middleware"
	"identity/repository"
	"identity/service"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup logger
	appLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Auto migrate
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.Role{},
		&domain.UserRole{},
	); err != nil {
		appLogger.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Seed default roles
	seedRoles(db)

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)

	// Setup services
	passwordSvc := service.NewPasswordService()
	tokenSvc := service.NewTokenService(&cfg.JWT)
	authSvc := service.NewAuthService(userRepo, refreshRepo, tokenSvc, passwordSvc, cfg, appLogger)

	// Setup handlers
	authHandler := handler.NewAuthHandler(authSvc, appLogger)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "identity"})
	})

	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.JWTAuth(&cfg.JWT))
	{
		protected.GET("/me", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"user_id":  c.GetString("user_id"),
				"email":    c.GetString("email"),
				"username": c.GetString("username"),
				"roles":    c.GetStringSlice("roles"),
			})
		})
	}

	// Admin routes
	admin := router.Group("/api/admin")
	admin.Use(middleware.JWTAuth(&cfg.JWT))
	admin.Use(middleware.RequireRole("Admin"))
	{
		// Add admin routes here
	}

	// Start server
	appLogger.Info("Starting identity service", "port", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		appLogger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func seedRoles(db *gorm.DB) {
	roles := []domain.Role{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Admin", Description: "Administrator"},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "User", Description: "Regular user"},
	}

	for _, role := range roles {
		if err := db.FirstOrCreate(&role, "id = ?", role.ID).Error; err != nil {
			slog.Error("Failed to seed role", "role", role.Name, "error", err)
		}
	}
}
