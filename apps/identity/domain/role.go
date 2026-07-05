package domain

import (
	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`

	UserRoles []UserRole `json:"-" gorm:"foreignKey:RoleID"`
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	RoleID uuid.UUID `json:"role_id" gorm:"type:uuid;primaryKey"`

	User User `json:"-" gorm:"foreignKey:UserID"`
	Role Role `json:"-" gorm:"foreignKey:RoleID"`
}

func (Role) TableName() string {
	return "roles"
}

func (UserRole) TableName() string {
	return "user_roles"
}
