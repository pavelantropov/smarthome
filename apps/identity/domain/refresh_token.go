package domain

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Token       string     `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt   time.Time  `json:"expires_at" gorm:"not null"`
	IsRevoked   bool       `json:"is_revoked" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	CreatedByIP string     `json:"created_by_ip"`
	RevokedAt   *time.Time `json:"revoked_at"`
	RevokedByIP string     `json:"revoked_by_ip"`
	Reason      string     `json:"reason"`

	User User `json:"-" gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (rt *RefreshToken) IsActive() bool {
	return !rt.IsRevoked && rt.ExpiresAt.After(time.Now())
}
