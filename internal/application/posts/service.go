package posts

import (
	"context"
	"strings"

	"blog_api/internal/application"
	"blog_api/internal/application/auth"
	"blog_api/internal/domain"
)

type Service struct {
	repo domain.PostRepository
}

func NewService(repo domain.PostRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, identity auth.Identity) ([]domain.Post, error) {
	if auth.CanManageAllPosts(identity) {
		return s.repo.List(ctx)
	}
	return s.repo.ListByAuthor(ctx, identity.UserID)
}

func (s *Service) Get(ctx context.Context, identity auth.Identity, id uint64) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, application.ErrNotFound
	}
	if err := auth.CanAccessPost(identity, post.AuthorID); err != nil {
		return nil, err
	}
	return post, nil
}

type UpsertInput struct {
	Title    string
	Content  string
	Status   domain.PostStatus
	AuthorID *uint64
}

func validateInput(input UpsertInput) error {
	if strings.TrimSpace(input.Title) == "" || len(input.Title) > 255 {
		return application.ErrValidation
	}
	if strings.TrimSpace(input.Content) == "" {
		return application.ErrValidation
	}
	if input.Status != domain.PostStatusDraft && input.Status != domain.PostStatusPublished {
		return application.ErrValidation
	}
	return nil
}

func (s *Service) Create(ctx context.Context, identity auth.Identity, input UpsertInput) (*domain.Post, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	authorID := identity.UserID
	if input.AuthorID != nil {
		authorID = *input.AuthorID
	}
	if err := auth.CanCreateForAuthor(identity, authorID); err != nil {
		return nil, err
	}

	post := &domain.Post{
		Title:    strings.TrimSpace(input.Title),
		Content:  strings.TrimSpace(input.Content),
		AuthorID: authorID,
		Status:   input.Status,
	}
	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *Service) Update(ctx context.Context, identity auth.Identity, id uint64, input UpsertInput) (*domain.Post, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, application.ErrNotFound
	}
	if err := auth.CanAccessPost(identity, post.AuthorID); err != nil {
		return nil, err
	}
	if input.AuthorID != nil && *input.AuthorID != post.AuthorID {
		if err := auth.CanCreateForAuthor(identity, *input.AuthorID); err != nil {
			return nil, err
		}
		post.AuthorID = *input.AuthorID
	}
	post.Title = strings.TrimSpace(input.Title)
	post.Content = strings.TrimSpace(input.Content)
	post.Status = input.Status
	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *Service) Delete(ctx context.Context, identity auth.Identity, id uint64) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post == nil {
		return application.ErrNotFound
	}
	if err := auth.CanAccessPost(identity, post.AuthorID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
