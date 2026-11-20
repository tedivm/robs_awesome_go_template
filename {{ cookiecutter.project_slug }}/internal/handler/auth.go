package handler

import (
	"context"
	"log/slog"
	"net/http"

	"{{ cookiecutter.import_path }}/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

// SetupAuth registers the auth endpoint.
func SetupAuth(api huma.API, logger *slog.Logger) {
	// POST /api/v1/auth/login — Login and get JWT token
	huma.Post(api, "/api/v1/auth/login", func(ctx context.Context, input *struct {
		Body struct {
			APIKey string `json:"api_key" doc:"API key for authentication"`
		}
	}) (*struct {
		Body struct {
			Token string `json:"token" doc:"JWT token"`
		}
	}, error) {
		// Validate the API key against the configured value.
		// When api_key is not set (development mode), any key is accepted.
		// In production, ValidateForProduction() ensures api_key is configured.
		if err := service.ValidateAPIKey(input.Body.APIKey); err != nil {
			return nil, huma.NewError(http.StatusUnauthorized, "invalid api key")
		}
		// Use a stable identifier as the token subject, not the raw API key.
		token, err := service.GenerateToken("service")
		if err != nil {
			logger.ErrorContext(ctx, "generate token failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct {
			Body struct {
				Token string `json:"token" doc:"JWT token"`
			}
		}{Body: struct {
			Token string `json:"token" doc:"JWT token"`
		}{Token: token}}, nil
	})
}
