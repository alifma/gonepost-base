package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port           string
	AllowedOrigins []string
	DatabaseURL    string
	// CookieSecure controls the Secure attribute on the session cookie.
	// Must be true in any real deployment (cookie only sent over HTTPS).
	// Only false when APP_ENV=local, so local HTTP testing (curl/Bruno)
	// doesn't have the cookie silently dropped.
	CookieSecure bool
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigins := []string{}
	cors := os.Getenv("CORS_ALLOWED_ORIGINS")
	if cors != "" {
		allowedOrigins = strings.Split(cors, ",")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return Config{
		Port:           ":" + port,
		AllowedOrigins: allowedOrigins,
		DatabaseURL:    databaseURL,
		CookieSecure:   os.Getenv("APP_ENV") != "local",
	}, nil
}
