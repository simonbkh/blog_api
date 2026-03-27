package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"blog_api/internal/application"
	"blog_api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users           domain.UserRepository
	refreshTokens   domain.RefreshTokenRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(
	users domain.UserRepository,
	refreshTokens domain.RefreshTokenRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		users:           users,
		refreshTokens:   refreshTokens,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func HashPassword(password string) (string, error) {
	if len(strings.TrimSpace(password)) < 8 {
		return "", application.ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *Service) EnsureSuperAdmin(ctx context.Context, email, password, fullName string) error {
	if email == "" || password == "" {
		return nil
	}
	existing, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	user := &domain.User{
		FullName:     strings.TrimSpace(fullName),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
		Role:         domain.RoleSuperAdmin,
	}
	return s.users.Create(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, Identity, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, Identity{}, err
	}
	if user == nil {
		return nil, Identity{}, application.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, Identity{}, application.ErrUnauthorized
	}

	identity := Identity{UserID: user.ID, Role: user.Role, Email: user.Email}
	pair, err := s.issueTokenPair(ctx, identity)
	if err != nil {
		return nil, Identity{}, err
	}

	return pair, identity, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := hashToken(refreshToken)
	session, err := s.refreshTokens.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session == nil || session.IsRevoked() || time.Now().After(session.ExpiresAt) {
		return nil, application.ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, application.ErrUnauthorized
	}

	identity := Identity{UserID: user.ID, Role: user.Role, Email: user.Email}
	newPair, newSession, err := s.buildTokenPair(identity)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.refreshTokens.Rotate(ctx, session.ID, newSession, now); err != nil {
		return nil, err
	}

	return newPair, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	session, err := s.refreshTokens.GetByTokenHash(ctx, hash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}
	now := time.Now().UTC()
	return s.refreshTokens.RevokeByTokenID(ctx, session.ID, now)
}

func (s *Service) ParseAccessToken(tokenString string) (Identity, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return Identity{}, application.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return Identity{}, application.ErrUnauthorized
	}
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return Identity{}, application.ErrUnauthorized
	}
	roleStr, ok := claims["role"].(string)
	if !ok {
		return Identity{}, application.ErrUnauthorized
	}
	emailStr, _ := claims["email"].(string)

	return Identity{UserID: uint64(userIDFloat), Role: domain.Role(roleStr), Email: emailStr}, nil
}

func (s *Service) issueTokenPair(ctx context.Context, identity Identity) (*TokenPair, error) {
	pair, session, err := s.buildTokenPair(identity)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Create(ctx, session); err != nil {
		return nil, err
	}
	return pair, nil
}

func (s *Service) buildTokenPair(identity Identity) (*TokenPair, *domain.RefreshTokenSession, error) {
	now := time.Now().UTC()
	accessExp := now.Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub":   identity.UserID,
		"role":  string(identity.Role),
		"email": identity.Email,
		"exp":   accessExp.Unix(),
		"iat":   now.Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := jwtToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, nil, err
	}

	rawRefresh, err := randomToken(48)
	if err != nil {
		return nil, nil, err
	}
	sessionID, err := randomToken(24)
	if err != nil {
		return nil, nil, err
	}
	refreshExp := now.Add(s.refreshTokenTTL)

	session := &domain.RefreshTokenSession{
		ID:        sessionID,
		UserID:    identity.UserID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: refreshExp,
	}

	pair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    accessExp,
	}
	return pair, session, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ParseBearerToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("empty bearer token")
	}
	return parts[1], nil
}

func ParseUserID(subject string) (uint64, error) {
	id, err := strconv.ParseUint(subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
