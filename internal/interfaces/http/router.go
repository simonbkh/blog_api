package httpapi

import (
	"blog_api/internal/application/auth"
	"blog_api/internal/application/posts"
	"blog_api/internal/config"
	"blog_api/internal/domain"
	"blog_api/internal/interfaces/http/handlers"
	"blog_api/internal/interfaces/http/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Dependencies struct {
	Config      config.Config
	AuthService *auth.Service
	PostService *posts.Service
	StartedAt   time.Time
}

func NewRouter(deps Dependencies) http.Handler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	authHandler := handlers.NewAuthHandler(deps.AuthService, validate)
	postsHandler := handlers.NewPostsHandler(deps.PostService, validate)
	systemHandler := handlers.NewSystemHandler(deps.StartedAt)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(20 * time.Second))
	r.Use(middleware.CORS(deps.Config.AllowedOrigins))
	r.Use(middleware.Authentication(deps.AuthService))
	r.Use(middleware.RateLimit(
		deps.Config.RateLimitIPPerMinute,
		deps.Config.RateLimitUserPerMinute,
		func(r *http.Request) (string, bool) {
			identity, ok := middleware.IdentityFromContext(r.Context())
			if !ok {
				return "", false
			}
			return strconv.FormatUint(identity.UserID, 10), true
		},
	))

	r.Get("/healthz", systemHandler.Health)
	// Serve minimal admin SPA files.
	r.Handle("/admin/*", http.StripPrefix("/admin/", http.FileServer(http.Dir("web/admin"))))

	r.Route("/auth", func(ar chi.Router) {
		ar.Post("/login", authHandler.Login)
		ar.Post("/refresh", authHandler.Refresh)
		ar.With(middleware.RequireAuth).Post("/logout", authHandler.Logout)
	})

	r.Group(func(pr chi.Router) {
		pr.Use(middleware.RequireAuth)
		pr.Get("/me", authHandler.Me)
		pr.Route("/posts", func(ps chi.Router) {
			ps.Get("/", postsHandler.List)
			ps.Post("/", postsHandler.Create)
			ps.Get("/{id}", postsHandler.Get)
			ps.Put("/{id}", postsHandler.Update)
			ps.Delete("/{id}", postsHandler.Delete)
		})
	})

	r.With(middleware.RequireAnyRole(domain.RoleSuperAdmin)).Get("/admin/system/health", systemHandler.Health)
	return r
}
