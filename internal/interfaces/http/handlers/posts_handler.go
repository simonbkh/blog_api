package handlers

import (
	"blog_api/internal/application"
	"blog_api/internal/application/posts"
	"blog_api/internal/domain"
	"blog_api/internal/interfaces/http/middleware"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type PostsHandler struct {
	svc      *posts.Service
	validate *validator.Validate
}

func NewPostsHandler(svc *posts.Service, validate *validator.Validate) *PostsHandler {
	return &PostsHandler{svc: svc, validate: validate}
}

type postRequest struct {
	Title    string            `json:"title" validate:"required,min=3,max=255"`
	Content  string            `json:"content" validate:"required,min=1"`
	Status   domain.PostStatus `json:"status" validate:"required,oneof=draft published"`
	AuthorID *uint64           `json:"author_id,omitempty"`
	ImageIDs []uint64          `json:"image_ids,omitempty"`
}

func (h *PostsHandler) List(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	postsList, err := h.svc.List(r.Context(), identity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, postsList)
}

func (h *PostsHandler) Get(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	post, err := h.svc.Get(r.Context(), identity, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (h *PostsHandler) Create(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	var req postRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	created, err := h.svc.Create(r.Context(), identity, posts.UpsertInput{
		Title:    req.Title,
		Content:  req.Content,
		Status:   req.Status,
		AuthorID: req.AuthorID,
		ImageIDs: req.ImageIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *PostsHandler) Update(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	var req postRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	updated, err := h.svc.Update(r.Context(), identity, id, posts.UpsertInput{
		Title:    req.Title,
		Content:  req.Content,
		Status:   req.Status,
		AuthorID: req.AuthorID,
		ImageIDs: req.ImageIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *PostsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.svc.Delete(r.Context(), identity, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(value string) (uint64, error) {
	return strconv.ParseUint(value, 10, 64)
}
