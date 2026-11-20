package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"{{ cookiecutter.import_path }}/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)

{% if cookiecutter.include_web == "y" %}
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Environment)
{% endif %}{% if cookiecutter.include_orm == "y" %}
	assert.Equal(t, "sqlite", cfg.Database.Driver)
{% endif %}{% if cookiecutter.include_workers == "y" %}
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
{% endif %}
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
{% if cookiecutter.include_web == "y" %}
	assert.Empty(t, cfg.Auth.JwtSecret)
	assert.Empty(t, cfg.Auth.APIKey)
	assert.Equal(t, "api_key", cfg.Auth.AuthMethod)
{% endif %}
}

func TestLoadFromEnv(t *testing.T) {
{% if cookiecutter.include_web == "y" %}
	require.NoError(t, os.Setenv("APP_SERVER_PORT", "9090"))
	require.NoError(t, os.Setenv("APP_SERVER_ENVIRONMENT", "production"))
	t.Cleanup(func() {
		os.Unsetenv("APP_SERVER_PORT")
		os.Unsetenv("APP_SERVER_ENVIRONMENT")
	})
{% endif %}
	require.NoError(t, os.Setenv("APP_LOG_LEVEL", "debug"))
	require.NoError(t, os.Setenv("APP_LOG_FORMAT", "json"))
	t.Cleanup(func() {
		os.Unsetenv("APP_LOG_LEVEL")
		os.Unsetenv("APP_LOG_FORMAT")
	})

	cfg, err := config.Load()
	require.NoError(t, err)

{% if cookiecutter.include_web == "y" %}
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "production", cfg.Server.Environment)
{% endif %}
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoadFromConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
{% if cookiecutter.include_web == "y" %}
	const yamlContent = `
server:
  port: 7070
  environment: test
{% if cookiecutter.include_orm == "y" %}
database:
  driver: postgres
  dsn: host=localhost dbname=test
{% endif %}log:
  level: trace
  format: json
`
{% else %}
	const yamlContent = `
{% if cookiecutter.include_orm == "y" %}
database:
  driver: postgres
  dsn: host=localhost dbname=test
{% endif %}log:
  level: trace
  format: json
`
{% endif %}
	require.NoError(t, os.WriteFile(configFile, []byte(yamlContent), 0644))

	require.NoError(t, os.Setenv("CONFIG_PATH", configFile))
	t.Cleanup(func() {
		os.Unsetenv("CONFIG_PATH")
	})

	cfg, err := config.Load()
	require.NoError(t, err)

{% if cookiecutter.include_web == "y" %}
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Environment)
{% endif %}{% if cookiecutter.include_orm == "y" %}
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "host=localhost dbname=test", cfg.Database.DSN)
{% endif %}
	assert.Equal(t, "trace", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestEnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
{% if cookiecutter.include_web == "y" %}
	const yamlContent = `
server:
  port: 7070
log:
  level: info
`
{% else %}
	const yamlContent = `
log:
  level: info
`
{% endif %}
	require.NoError(t, os.WriteFile(configFile, []byte(yamlContent), 0644))

	require.NoError(t, os.Setenv("CONFIG_PATH", configFile))
{% if cookiecutter.include_web == "y" %}
	require.NoError(t, os.Setenv("APP_SERVER_PORT", "9090"))
	t.Cleanup(func() {
		os.Unsetenv("CONFIG_PATH")
		os.Unsetenv("APP_SERVER_PORT")
	})

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
{% else %}
	require.NoError(t, os.Setenv("APP_LOG_LEVEL", "debug"))
	t.Cleanup(func() {
		os.Unsetenv("CONFIG_PATH")
		os.Unsetenv("APP_LOG_LEVEL")
	})

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.Log.Level)
{% endif %}
}

func TestMissingConfigFileIsOk(t *testing.T) {
	require.NoError(t, os.Setenv("CONFIG_PATH", "/nonexistent/config.yaml"))
	t.Cleanup(func() {
		os.Unsetenv("CONFIG_PATH")
	})

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "info", cfg.Log.Level)
{% if cookiecutter.include_web == "y" %}
	assert.Equal(t, 8080, cfg.Server.Port)
{% endif %}{% if cookiecutter.include_orm == "y" %}
	assert.Equal(t, "sqlite", cfg.Database.Driver)
{% endif %}
}

func TestEnvVarOverride(t *testing.T) {
{% if cookiecutter.include_web == "y" %}
	t.Setenv("APP_SERVER__PORT", "9090")
{% endif %}
	t.Setenv("APP_LOG__LEVEL", "warn")

	cfg, err := config.Load()
	require.NoError(t, err)

{% if cookiecutter.include_web == "y" %}
	assert.Equal(t, 9090, cfg.Server.Port)
{% endif %}
	assert.Equal(t, "warn", cfg.Log.Level)
{% if cookiecutter.include_orm == "y" %}
	assert.Equal(t, "sqlite", cfg.Database.Driver)
{% endif %}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := &config.Config{
{% if cookiecutter.include_web == "y" %}
		Server: config.ServerConfig{
			Port:         8080,
			Environment:  "development",
			ReadTimeout:  30,
			WriteTimeout: 30,
			RateLimit:    100,
		},
{% endif %}{% if cookiecutter.include_orm == "y" %}
		Database: config.DatabaseConfig{
			Driver:          "sqlite",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
		},
{% endif %}
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
		},
{% if cookiecutter.include_web == "y" %}
		Auth: config.AuthConfig{AuthMethod: "api_key", JwtSecret: "test-secret", APIKey: "test-api-key"},
		Cors: config.CorsConfig{
			AllowedOrigins:  []string{"*"},
			AllowedMethods:  []string{"GET", "POST"},
			AllowedHeaders:  []string{"Content-Type"},
		},
{% endif %}
	}

	assert.NoError(t, cfg.Validate())
}

func TestValidateInvalidLogLevel(t *testing.T) {
	cfg := &config.Config{
{% if cookiecutter.include_web == "y" %}
		Server: config.ServerConfig{Port: 8080},
{% endif %}{% if cookiecutter.include_orm == "y" %}
		Database: config.DatabaseConfig{Driver: "sqlite"},
{% endif %}
		Log: config.LogConfig{Level: "invalid", Format: "text"},
{% if cookiecutter.include_web == "y" %}
		Auth: config.AuthConfig{AuthMethod: "api_key", JwtSecret: "test-secret", APIKey: "test-api-key"},
{% endif %}
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "level")
}

{% if cookiecutter.include_orm == "y" %}
func TestValidatePostgresWithoutDSN(t *testing.T) {
	cfg := &config.Config{
{% if cookiecutter.include_web == "y" %}
		Server: config.ServerConfig{Port: 8080},
{% endif %}
		Database: config.DatabaseConfig{
			Driver: "postgres",
		},
		Log: config.LogConfig{Level: "info", Format: "text"},
{% if cookiecutter.include_web == "y" %}
		Auth: config.AuthConfig{AuthMethod: "api_key", JwtSecret: "test-secret", APIKey: "test-api-key"},
{% endif %}
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DSN")
}

func TestValidateUnsupportedDriver(t *testing.T) {
	cfg := &config.Config{
{% if cookiecutter.include_web == "y" %}
		Server: config.ServerConfig{Port: 8080},
{% endif %}
		Database: config.DatabaseConfig{Driver: "mysql"},
		Log: config.LogConfig{Level: "info", Format: "text"},
{% if cookiecutter.include_web == "y" %}
		Auth: config.AuthConfig{AuthMethod: "api_key", JwtSecret: "test-secret", APIKey: "test-api-key"},
{% endif %}
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver")
}
{% endif %}