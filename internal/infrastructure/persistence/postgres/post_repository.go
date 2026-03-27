package postgres

import (
	"context"
	"errors"

	"blog_api/internal/domain"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) List(ctx context.Context) ([]domain.Post, error) {
	var models []PostModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return mapPosts(models), nil
}

func (r *PostRepository) ListByAuthor(ctx context.Context, authorID uint64) ([]domain.Post, error) {
	var models []PostModel
	if err := r.db.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return mapPosts(models), nil
}

func (r *PostRepository) GetByID(ctx context.Context, id uint64) (*domain.Post, error) {
	var model PostModel
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapPostDomain(model), nil
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	model := PostModel{
		Title:    post.Title,
		Content:  post.Content,
		AuthorID: post.AuthorID,
		Status:   string(post.Status),
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	post.ID = model.ID
	post.CreatedAt = model.CreatedAt
	post.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	update := map[string]any{
		"title":     post.Title,
		"content":   post.Content,
		"author_id": post.AuthorID,
		"status":    string(post.Status),
	}
	if err := r.db.WithContext(ctx).Model(&PostModel{}).Where("id = ?", post.ID).Updates(update).Error; err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).Delete(&PostModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func mapPosts(models []PostModel) []domain.Post {
	out := make([]domain.Post, 0, len(models))
	for _, model := range models {
		out = append(out, *mapPostDomain(model))
	}
	return out
}

func mapPostDomain(model PostModel) *domain.Post {
	return &domain.Post{
		ID:        model.ID,
		Title:     model.Title,
		Content:   model.Content,
		AuthorID:  model.AuthorID,
		Status:    domain.PostStatus(model.Status),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
