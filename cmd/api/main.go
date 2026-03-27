package main

import (
	"blog_api/internal/application/auth"
	"blog_api/internal/application/posts"
	"blog_api/internal/config"
	"blog_api/internal/infrastructure/persistence/postgres"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "blog_api/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := postgres.New(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db init error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := postgres.Ping(ctx, db); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	if err := postgres.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	postRepo := postgres.NewPostRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)

	authService := auth.NewService(userRepo, refreshRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	if err := authService.EnsureSuperAdmin(context.Background(), cfg.SeedSuperAdminEmail, cfg.SeedSuperAdminPassword, cfg.SeedSuperAdminFullName); err != nil {
		log.Fatalf("seed super admin error: %v", err)
	}
	postService := posts.NewService(postRepo)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Config:      cfg,
		AuthService: authService,
		PostService: postService,
		StartedAt:   time.Now().UTC(),
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
}
