package handlers

import (
	"net/http"
	"strings"
	"time"

	"blog_api/internal/application"
	"blog_api/internal/application/auth"
	"blog_api/internal/interfaces/http/middleware"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authSvc  *auth.Service
	validate *validator.Validate
}

func NewAuthHandler(authSvc *auth.Service, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, validate: validate}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	pair, identity, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tokens": tokenResponse{
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			ExpiresAt:    pair.ExpiresAt,
		},
		"user": map[string]any{
			"id":    identity.UserID,
			"email": identity.Email,
			"role":  identity.Role,
		},
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	pair, err := h.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, ExpiresAt: pair.ExpiresAt})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		writeError(w, application.ErrValidation)
		return
	}
	if err := h.authSvc.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    identity.UserID,
		"email": identity.Email,
		"role":  identity.Role,
	})
}
