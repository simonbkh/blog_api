package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	HTTPAddr               string
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	IdleTimeout            time.Duration
	ShutdownTimeout        time.Duration
	DBDSN                  string
	JWTSecret              string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	AllowedOrigins         []string
	RateLimitIPPerMinute   int
	RateLimitUserPerMinute int
	SeedSuperAdminEmail    string
	SeedSuperAdminPassword string
	SeedSuperAdminFullName string
}

func Load() (Config, error) {
	loadEnv()

	cfg := Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		HTTPAddr:               getEnv("HTTP_ADDR", ":8080"),
		ReadTimeout:            getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:           getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:            getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:        getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		DBDSN:                  getEnv("DB_DSN", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		AccessTokenTTL:         getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:        getEnvDuration("REFRESH_TOKEN_TTL", 24*time.Hour*7),
		AllowedOrigins:         splitCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8080")),
		RateLimitIPPerMinute:   getEnvInt("RATE_LIMIT_IP_PER_MINUTE", 120),
		RateLimitUserPerMinute: getEnvInt("RATE_LIMIT_USER_PER_MINUTE", 300),
		SeedSuperAdminEmail:    getEnv("SEED_SUPERADMIN_EMAIL", ""),
		SeedSuperAdminPassword: getEnv("SEED_SUPERADMIN_PASSWORD", ""),
		SeedSuperAdminFullName: getEnv("SEED_SUPERADMIN_FULL_NAME", "Super Admin"),
	}

	if cfg.DBDSN == "" {
		return cfg, fmt.Errorf("DB_DSN is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET must be at least 32 chars")
	}
	if cfg.RateLimitIPPerMinute < 1 || cfg.RateLimitUserPerMinute < 1 {
		return cfg, fmt.Errorf("rate limits must be greater than zero")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return cfg, fmt.Errorf("at least one ALLOWED_ORIGINS value is required")
	}

	return cfg, nil
}

func loadEnv() {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" || env == "development" || env == "local" {
		// In local development, prioritize .env values to avoid stale shell exports.
		_ = godotenv.Overload(".env")
		return
	}

	// In non-local environments, only load .env when variables are missing.
	_ = godotenv.Load(".env")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(val string) []string {
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
