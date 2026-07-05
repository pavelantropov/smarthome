package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email           string     `json:"email" gorm:"uniqueIndex;not null"`
	Username        string     `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash    string     `json:"-" gorm:"not null"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	PhoneNumber     string     `json:"phone_number"`
	IsEmailVerified bool       `json:"is_email_verified" gorm:"default:false"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	LastLoginAt     *time.Time `json:"last_login_at"`

	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
	UserRoles     []UserRole     `json:"-" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}
