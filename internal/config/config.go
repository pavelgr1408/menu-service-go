package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	HTTPPort        string
	DatabaseURL     string
	DatabaseHost    string
	DatabasePort    string
	DatabaseName    string
	DatabaseUser    string
	DatabasePass    string
	DatabaseSSLMode string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:        env("HTTP_PORT", "8080"),
		DatabaseHost:    env("DATABASE_HOST", "localhost"),
		DatabasePort:    env("DATABASE_PORT", "15434"),
		DatabaseName:    env("DATABASE_NAME", "restaurant_menu_db"),
		DatabaseUser:    env("DATABASE_USER", "restaurant_menu_user"),
		DatabasePass:    env("DATABASE_PASSWORD", "restaurant_menu_password"),
		DatabaseSSLMode: env("DATABASE_SSLMODE", "disable"),
	}
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		cfg.DatabaseURL = raw
		return cfg, nil
	}
	cfg.DatabaseURL = buildDatabaseURL(cfg)
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildDatabaseURL(cfg Config) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DatabaseUser, cfg.DatabasePass),
		Host:   fmt.Sprintf("%s:%s", cfg.DatabaseHost, cfg.DatabasePort),
		Path:   cfg.DatabaseName,
	}
	q := u.Query()
	q.Set("sslmode", cfg.DatabaseSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
