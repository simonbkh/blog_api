package postgres

import (
	"context"
	"errors"
	"sort"

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
	return r.attachImages(ctx, mapPosts(models))
}

func (r *PostRepository) ListByAuthor(ctx context.Context, authorID uint64) ([]domain.Post, error) {
	var models []PostModel
	if err := r.db.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return r.attachImages(ctx, mapPosts(models))
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
	post := mapPostDomain(model)
	images, err := r.loadImageIDs(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Images = images
	return post, nil
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
	return r.syncImages(ctx, post.ID, post.Images)
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
	return r.syncImages(ctx, post.ID, post.Images)
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

// syncImages replaces the post_images rows for a given post.
func (r *PostRepository) syncImages(ctx context.Context, postID uint64, mediaIDs []uint64) error {
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&PostImageModel{}).Error; err != nil {
		return err
	}
	for i, mid := range mediaIDs {
		row := PostImageModel{PostID: postID, MediaID: mid, Position: i}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

// loadImageIDs returns media IDs for a single post, ordered by position.
func (r *PostRepository) loadImageIDs(ctx context.Context, postID uint64) ([]uint64, error) {
	var rows []PostImageModel
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).Order("position").Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MediaID)
	}
	return ids, nil
}

// attachImages bulk-loads images for a slice of posts.
func (r *PostRepository) attachImages(ctx context.Context, posts []domain.Post) ([]domain.Post, error) {
	if len(posts) == 0 {
		return posts, nil
	}
	postIDs := make([]uint64, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	var rows []PostImageModel
	if err := r.db.WithContext(ctx).Where("post_id IN ?", postIDs).Order("position").Find(&rows).Error; err != nil {
		return nil, err
	}
	mapping := make(map[uint64][]uint64)
	for _, row := range rows {
		mapping[row.PostID] = append(mapping[row.PostID], row.MediaID)
	}
	for i := range posts {
		posts[i].Images = mapping[posts[i].ID]
	}
	return posts, nil
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

// Ensure sort import is used.
var _ = sort.Strings
