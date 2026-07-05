package repository

import (
	"context"
	"errors"
	"identity/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role").
		First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role").
		First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role").
		First(&user, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role").
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) IsEmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.New()
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	userRole := &domain.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}
