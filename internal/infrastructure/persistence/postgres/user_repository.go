package postgres

import (
	"context"
	"errors"
	"strings"

	"blog_api/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapUserDomain(model), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", strings.ToLower(strings.TrimSpace(email))).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapUserDomain(model), nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	model := UserModel{
		FullName:     user.FullName,
		Email:        strings.ToLower(strings.TrimSpace(user.Email)),
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func mapUserDomain(model UserModel) *domain.User {
	return &domain.User{
		ID:           model.ID,
		FullName:     model.FullName,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Role:         domain.Role(model.Role),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
