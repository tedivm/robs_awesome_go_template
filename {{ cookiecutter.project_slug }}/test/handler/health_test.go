package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"{{ cookiecutter.import_path }}/internal/handler"
	"{{ cookiecutter.import_path }}/internal/model"
	"{{ cookiecutter.import_path }}/internal/service"
	"{{ cookiecutter.import_path }}/test/util"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHealthTest(t *testing.T) *chi.Mux {
	t.Helper()

	db := util.TestDB(t)
	util.CreateTables(t, db, &model.Item{})

	service.Init(db, slog.Default(), "test-secret", "test-api-key")
	t.Cleanup(func() { _ = service.Close() })

	r := chi.NewRouter()
	r.Get("/health", handler.HealthCheck(db, slog.Default()))

	return r
}

func TestHealthCheck(t *testing.T) {

	r := setupHealthTest(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "ok", result["status"])
}

func TestHealthCheckNilDB(t *testing.T) {

	r := chi.NewRouter()
	r.Get("/health", handler.HealthCheck(nil, slog.Default()))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
