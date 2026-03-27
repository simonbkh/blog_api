package postgres

import (
	"context"
	"errors"
	"time"

	"blog_api/internal/domain"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshTokenSession) error {
	model := RefreshTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	token.CreatedAt = model.CreatedAt
	return nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshTokenSession, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapRefreshDomain(model), nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldTokenID string, newToken *domain.RefreshTokenSession, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&RefreshTokenModel{}).Where("id = ?", oldTokenID).Updates(map[string]any{
			"revoked_at":  revokedAt,
			"replaced_by": newToken.ID,
		}).Error; err != nil {
			return err
		}
		model := RefreshTokenModel{
			ID:        newToken.ID,
			UserID:    newToken.UserID,
			TokenHash: newToken.TokenHash,
			ExpiresAt: newToken.ExpiresAt,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		newToken.CreatedAt = model.CreatedAt
		return nil
	})
}

func (r *RefreshTokenRepository) RevokeByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("id = ?", tokenID).Update("revoked_at", revokedAt).Error
}

func mapRefreshDomain(model RefreshTokenModel) *domain.RefreshTokenSession {
	return &domain.RefreshTokenSession{
		ID:         model.ID,
		UserID:     model.UserID,
		TokenHash:  model.TokenHash,
		ExpiresAt:  model.ExpiresAt,
		RevokedAt:  model.RevokedAt,
		ReplacedBy: model.ReplacedBy,
		CreatedAt:  model.CreatedAt,
	}
}
