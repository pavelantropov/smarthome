package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey         string
	Issuer            string
	Audience          string
	ExpiryMinutes     time.Duration
	RefreshExpiryDays time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "identity_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey:         getEnv("JWT_SECRET", "your-256-bit-secret-key-change-in-production"),
			Issuer:            getEnv("JWT_ISSUER", "SmartHome.Identity"),
			Audience:          getEnv("JWT_AUDIENCE", "SmartHome.Clients"),
			ExpiryMinutes:     time.Duration(getEnvAsInt("JWT_EXPIRY_MINUTES", 15)) * time.Minute,
			RefreshExpiryDays: time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 7)) * 24 * time.Hour,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
