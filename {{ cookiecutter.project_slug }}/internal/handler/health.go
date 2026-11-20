package handler

import (
	"net/http"
{% if cookiecutter.include_orm == "y" %}
	"encoding/json"
	"log/slog"

	"github.com/uptrace/bun"
{% else %}
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
{% endif %}
)

// HealthCheck returns an HTTP handler that reports service health.
{% if cookiecutter.include_orm == "y" %}
// It checks database connectivity if a database is provided.
func HealthCheck(db *bun.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]string{"status": "ok"}

		// Check database connectivity
		if db != nil {
			if err := db.Ping(); err != nil {
				health["status"] = "error"
				health["database"] = "connection failed"
				// Log the real error for admins — don't leak it to the client.
				logger.ErrorContext(r.Context(), "health check database ping failed", "error", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}
{% else %}
// It returns a simple OK response.
func HealthCheck(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// Setup is a no-op when ORM is disabled (routes are defined in items.go when ORM is present).
func Setup(api huma.API, logger *slog.Logger) {
}
{% endif %}
