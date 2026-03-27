package auth

import (
	"blog_api/internal/application"
	"blog_api/internal/domain"
)

func CanManageAllPosts(identity Identity) bool {
	return identity.Role.IsAdminLike()
}

func CanAccessPost(identity Identity, authorID uint64) error {
	if identity.Role.IsAdminLike() || identity.UserID == authorID {
		return nil
	}
	return application.ErrForbidden
}

func CanCreateForAuthor(identity Identity, authorID uint64) error {
	if identity.Role.IsAdminLike() || identity.UserID == authorID {
		return nil
	}
	return application.ErrForbidden
}

func RequireRoles(identity Identity, allowed ...domain.Role) error {
	for _, role := range allowed {
		if identity.Role == role {
			return nil
		}
	}
	return application.ErrForbidden
}
