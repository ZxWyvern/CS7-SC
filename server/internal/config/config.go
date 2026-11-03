package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr             string
	DatabaseURL      string
	AppOrigin        string
	CookieName       string
	SessionTTL       time.Duration
	CookieSecure     bool
	TrustProxy       bool
	ArgonMemory      uint32
	ArgonIterations  uint32
	ArgonParallelism uint8
}

func Load() (Config, error) {
	cfg := Config{
		Addr:             env("LISTEN_ADDR", ":8080"),
		DatabaseURL:      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AppOrigin:        strings.TrimRight(strings.TrimSpace(os.Getenv("APP_ORIGIN")), "/"),
		CookieName:       env("COOKIE_NAME", "cs7_session"),
		SessionTTL:       7 * 24 * time.Hour,
		ArgonMemory:      19456,
		ArgonIterations:  2,
		ArgonParallelism: 1,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL wajib diisi")
	}
	if cfg.AppOrigin == "" {
		return Config{}, errors.New("APP_ORIGIN wajib diisi")
	}

	origin, err := url.Parse(cfg.AppOrigin)
	if err != nil || origin.Scheme == "" || origin.Host == "" {
		return Config{}, fmt.Errorf("APP_ORIGIN tidak valid: %q", cfg.AppOrigin)
	}
	if origin.Scheme != "https" && origin.Scheme != "http" {
		return Config{}, errors.New("APP_ORIGIN harus memakai http atau https")
	}
	if origin.User != nil || (origin.Path != "" && origin.Path != "/") || origin.RawQuery != "" || origin.Fragment != "" {
		return Config{}, errors.New("APP_ORIGIN hanya boleh berisi skema dan host")
	}

	cookieSecure, err := boolEnv("COOKIE_SECURE", origin.Scheme == "https")
	if err != nil {
		return Config{}, err
	}
	cfg.CookieSecure = cookieSecure

	trustProxy, err := boolEnv("TRUST_PROXY", false)
	if err != nil {
		return Config{}, err
	}
	cfg.TrustProxy = trustProxy

	if raw := os.Getenv("SESSION_TTL"); raw != "" {
		ttl, err := time.ParseDuration(raw)
		if err != nil || ttl < time.Hour || ttl > 30*24*time.Hour {
			return Config{}, errors.New("SESSION_TTL harus antara 1h dan 720h")
		}
		cfg.SessionTTL = ttl
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func boolEnv(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s harus true atau false", key)
	}
	return value, nil
}
