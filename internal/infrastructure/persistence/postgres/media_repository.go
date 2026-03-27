package postgres

import (
	"context"
	"errors"

	"blog_api/internal/domain"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, media *domain.Media) error {
	model := MediaModel{
		OriginalName: media.OriginalName,
		StoredName:   media.StoredName,
		ContentType:  media.ContentType,
		SizeBytes:    media.SizeBytes,
		UploaderID:   media.UploaderID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	media.ID = model.ID
	media.CreatedAt = model.CreatedAt
	return nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id uint64) (*domain.Media, error) {
	var model MediaModel
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapMediaDomain(model), nil
}

func (r *MediaRepository) ListByIDs(ctx context.Context, ids []uint64) ([]domain.Media, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var models []MediaModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Media, 0, len(models))
	for _, m := range models {
		out = append(out, *mapMediaDomain(m))
	}
	return out, nil
}

func (r *MediaRepository) Delete(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).Delete(&MediaModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func mapMediaDomain(m MediaModel) *domain.Media {
	return &domain.Media{
		ID:           m.ID,
		OriginalName: m.OriginalName,
		StoredName:   m.StoredName,
		ContentType:  m.ContentType,
		SizeBytes:    m.SizeBytes,
		UploaderID:   m.UploaderID,
		CreatedAt:    m.CreatedAt,
	}
}
