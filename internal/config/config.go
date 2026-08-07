package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort            = 8000
	defaultShutdownTimeout = 10 * time.Second
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 20 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultSessionTTL      = 24 * time.Hour
)

type Config struct {
	Port            int
	DatabaseHost    string
	DatabasePort    int
	DatabaseUser    string
	DatabasePass    string
	DatabaseName    string
	DatabaseSSLMode string
	RedisURL        string
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	SessionTTL      time.Duration
	CookieSecure    bool
}

func Load() (Config, error) {
	cfg := Config{
		Port:            defaultPort,
		DatabaseHost:    valueOrDefault("DB_HOST", "localhost"),
		DatabasePort:    5432,
		DatabaseUser:    valueOrDefault("DB_USER", os.Getenv("POSTGRES_USER")),
		DatabasePass:    valueOrDefault("DB_PASSWORD", os.Getenv("POSTGRES_PASSWORD")),
		DatabaseName:    valueOrDefault("DB_NAME", os.Getenv("POSTGRES_DATABASE")),
		DatabaseSSLMode: valueOrDefault("DB_SSLMODE", "disable"),
		RedisURL:        valueOrDefault("REDIS_URL", "redis://localhost:6379"),
		AllowedOrigins:  splitCSV(valueOrDefault("CORS_ORIGINS", "http://localhost,http://localhost:5173")),
		ShutdownTimeout: defaultShutdownTimeout,
		ReadTimeout:     defaultReadTimeout,
		WriteTimeout:    defaultWriteTimeout,
		IdleTimeout:     defaultIdleTimeout,
		SessionTTL:      defaultSessionTTL,
	}

	var err error
	if cfg.Port, err = intValue("PORT", defaultPort); err != nil {
		return Config{}, err
	}
	if cfg.DatabasePort, err = intValue("DB_PORT", 5432); err != nil {
		return Config{}, err
	}
	if cfg.SessionTTL, err = durationValue("SESSION_TTL", defaultSessionTTL); err != nil {
		return Config{}, err
	}
	if cfg.CookieSecure, err = boolValue("COOKIE_SECURE", false); err != nil {
		return Config{}, err
	}

	missing := make([]string, 0, 3)
	if cfg.DatabaseUser == "" {
		missing = append(missing, "DB_USER/POSTGRES_USER")
	}
	if cfg.DatabasePass == "" {
		missing = append(missing, "DB_PASSWORD/POSTGRES_PASSWORD")
	}
	if cfg.DatabaseName == "" {
		missing = append(missing, "DB_NAME/POSTGRES_DATABASE")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func (c Config) HTTPAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c Config) DatabaseURL() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DatabaseUser, c.DatabasePass),
		Host:   fmt.Sprintf("%s:%d", c.DatabaseHost, c.DatabasePort),
		Path:   c.DatabaseName,
	}
	query := u.Query()
	query.Set("sslmode", c.DatabaseSSLMode)
	u.RawQuery = query.Encode()
	return u.String()
}

func intValue(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func durationValue(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return value, nil
}

func boolValue(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return value, nil
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
