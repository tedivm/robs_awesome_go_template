package config

import (
	"fmt"
	"os"
	"strings"
{% if cookiecutter.include_orm == "y" or cookiecutter.include_web == "y" %}
	"errors"
{% endif %}

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds the application configuration.
type Config struct {
{% if cookiecutter.include_web == "y" %}
	Server   ServerConfig   `koanf:"server"`
{% endif %}{% if cookiecutter.include_orm == "y" %}
	Database DatabaseConfig `koanf:"database"`
{% endif %}{% if cookiecutter.include_workers == "y" %}
	Redis    RedisConfig    `koanf:"redis"`
{% endif %}
	Log      LogConfig      `koanf:"log"`
{% if cookiecutter.include_web == "y" %}
	Auth     AuthConfig     `koanf:"auth"`
	Cors     CorsConfig     `koanf:"cors"`
{% endif %}
}

{% if cookiecutter.include_web == "y" %}
// ServerConfig holds server configuration.
type ServerConfig struct {
	Port         int    `koanf:"port"`
	Environment  string `koanf:"environment"`
	ReadTimeout  int    `koanf:"read_timeout"`
	WriteTimeout int    `koanf:"write_timeout"`
	RateLimit    int    `koanf:"rate_limit"`
}

{% endif %}{% if cookiecutter.include_orm == "y" %}
// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Driver          string `koanf:"driver"`
	DSN             string `koanf:"dsn"`
	MaxOpenConns    int    `koanf:"max_open_conns"`
	MaxIdleConns    int    `koanf:"max_idle_conns"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime"`
}

{% endif %}{% if cookiecutter.include_workers == "y" %}
// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addr     string `koanf:"addr"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

{% endif %}// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"` // "json" or "text"
}

{% if cookiecutter.include_web == "y" %}
// AuthConfig holds authentication configuration.
// AuthMethod controls the authentication mode:
//   - "none": authentication is disabled (all requests pass through)
//   - "api_key": requires a valid API key to obtain JWT tokens (default)
type AuthConfig struct {
	AuthMethod string `koanf:"auth_method"`
	JwtSecret  string `koanf:"jwt_secret"`
	APIKey     string `koanf:"api_key"`
}

// CorsConfig holds CORS configuration.
type CorsConfig struct {
	AllowedOrigins []string `koanf:"allowed_origins"`
	AllowedMethods []string `koanf:"allowed_methods"`
	AllowedHeaders []string `koanf:"allowed_headers"`
}
{% endif %}

// Validate checks the configuration for common misconfigurations.
func (c *Config) Validate() error {
{% if cookiecutter.include_web == "y" %}
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Server.ReadTimeout < 0 {
		return errors.New("server read_timeout cannot be negative")
	}
	if c.Server.WriteTimeout < 0 {
		return errors.New("server write_timeout cannot be negative")
	}
	if c.Server.RateLimit < 0 {
		return errors.New("server rate_limit cannot be negative")
	}

{% endif %}{% if cookiecutter.include_orm == "y" %}
	// Database validation
	switch c.Database.Driver {
	case "postgres":
		if c.Database.DSN == "" {
			return errors.New("postgres DSN is required when using postgres driver")
		}
	case "sqlite":
		// SQLite can work without DSN (uses file:memory: by default)
	default:
		return fmt.Errorf("unsupported database driver: %q", c.Database.Driver)
	}
	if c.Database.MaxOpenConns < 0 {
		return errors.New("database max_open_conns cannot be negative")
	}
	if c.Database.MaxIdleConns < 0 {
		return errors.New("database max_idle_conns cannot be negative")
	}
	if c.Database.ConnMaxLifetime < 0 {
		return errors.New("database conn_max_lifetime cannot be negative")
	}

{% endif %}	// Log validation
	switch c.Log.Level {
	case "debug", "info", "warn", "error", "trace":
		// Valid levels
	default:
		return fmt.Errorf("unsupported log level: %q", c.Log.Level)
	}
	switch c.Log.Format {
	case "json", "text":
		// Valid formats
	default:
		return fmt.Errorf("unsupported log format: %q", c.Log.Format)
	}

	return nil
}

// ValidateForProduction checks configuration for production readiness.
func (c *Config) ValidateForProduction() error {
	if err := c.Validate(); err != nil {
		return err
	}
{% if cookiecutter.include_orm == "y" %}

	// Require explicit DSN — don't allow defaults in production
	if c.Database.Driver == "sqlite" && c.Database.DSN == "app.db" {
		return errors.New("sqlite DSN should be explicitly set in production (default is CWD-relative)")
	}
	if c.Database.Driver == "postgres" && c.Database.DSN == "" {
		return errors.New("postgres DSN must be set in production")
	}
{% endif %}{% if cookiecutter.include_web == "y" %}

	// Auth validation: if auth is enabled, secrets must be set
	if c.Auth.AuthMethod != "none" {
		if c.Auth.JwtSecret == "" {
			return errors.New("jwt_secret must be set when auth is enabled")
		}
		if c.Auth.APIKey == "" {
			return errors.New("api_key must be set when auth is enabled")
		}
	}

	// Require non-default port
	if c.Server.Port == 8080 && c.Server.Environment == "production" {
		return errors.New("server port should be explicitly set in production")
	}
{% endif %}
	return nil
}

// Load loads configuration from file and environment variables.
// Config file paths are searched in order: explicit path via CONFIG_PATH env var,
// then configs/config.yaml relative to working directory.
func Load() (*Config, error) {
	k := koanf.New(".")

	// Load YAML config file (optional) — check CONFIG_PATH first, then default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Load .env file into os.Env (optional — if it exists, its values populate environment).
	// This allows .env to override YAML config via the env var loader below.
	// godotenv.IgnoreMissing() prevents error if .env doesn't exist.
	_ = godotenv.Load()

	// Load environment variables with APP_ prefix.
	// Use __ (double underscore) for nesting: APP_DATABASE__MAX_OPEN_CONNS → database.max_open_conns
	// Section part (before __) has underscores converted to dots. Field part (after __) is preserved.
	if err := k.Load(env.Provider("APP_", ".", func(s string) string {
		key := strings.TrimPrefix(s, "APP_")
		if idx := strings.Index(key, "__"); idx != -1 {
			section := strings.ToLower(strings.ReplaceAll(key[:idx], "_", "."))
			field := strings.ToLower(key[idx+2:])
			return section + "." + field
		}
		return strings.ToLower(strings.ReplaceAll(key, "_", "."))
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// Apply defaults
	applyDefaults(k)

	// Unmarshal into config struct
	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func applyDefaults(k *koanf.Koanf) {
{% if cookiecutter.include_web == "y" %}
	// Server defaults
	if !k.Exists("server.port") {
		k.Set("server.port", 8080)
	}
	if !k.Exists("server.environment") {
		k.Set("server.environment", "development")
	}
	if !k.Exists("server.read_timeout") {
		k.Set("server.read_timeout", 30)
	}
	if !k.Exists("server.write_timeout") {
		k.Set("server.write_timeout", 30)
	}
	if !k.Exists("server.rate_limit") {
		k.Set("server.rate_limit", 100)
	}

{% endif %}{% if cookiecutter.include_orm == "y" %}
	// Database defaults
	if !k.Exists("database.driver") {
		k.Set("database.driver", "sqlite")
	}
	if !k.Exists("database.dsn") {
		k.Set("database.dsn", "app.db")
	}
	if !k.Exists("database.max_open_conns") {
		k.Set("database.max_open_conns", 25)
	}
	if !k.Exists("database.max_idle_conns") {
		k.Set("database.max_idle_conns", 5)
	}
	if !k.Exists("database.conn_max_lifetime") {
		k.Set("database.conn_max_lifetime", 3600)
	}

{% endif %}{% if cookiecutter.include_workers == "y" %}
	// Redis defaults
	if !k.Exists("redis.addr") {
		k.Set("redis.addr", "localhost:6379")
	}
	if !k.Exists("redis.db") {
		k.Set("redis.db", 0)
	}

{% endif %}	// Log defaults
	if !k.Exists("log.level") {
		k.Set("log.level", "info")
	}
	if !k.Exists("log.format") {
		k.Set("log.format", "json")
	}
{% if cookiecutter.include_web == "y" %}

	// Auth defaults
	if !k.Exists("auth.auth_method") {
		k.Set("auth.auth_method", "api_key")
	}
	if !k.Exists("auth.jwt_secret") {
		k.Set("auth.jwt_secret", "")
	}
	if !k.Exists("auth.api_key") {
		k.Set("auth.api_key", "")
	}

	// CORS defaults
	if !k.Exists("cors.allowed_origins") {
		k.Set("cors.allowed_origins", []string{})
	}
	if !k.Exists("cors.allowed_methods") {
		k.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	}
	if !k.Exists("cors.allowed_headers") {
		k.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	}
{% endif %}
}
