package middleware

import (
	"context"
	"net/http"

	"blog_api/internal/application/auth"
)

type contextKey string

const identityContextKey contextKey = "identity"

type AuthParser interface {
	ParseAccessToken(tokenString string) (auth.Identity, error)
}

func Authentication(parser AuthParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header != "" {
				token, err := auth.ParseBearerToken(header)
				if err == nil {
					identity, parseErr := parser.ParseAccessToken(token)
					if parseErr == nil {
						r = r.WithContext(WithIdentity(r.Context(), identity))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := IdentityFromContext(r.Context()); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WithIdentity(ctx context.Context, identity auth.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

func IdentityFromContext(ctx context.Context) (auth.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey).(auth.Identity)
	return identity, ok
}
