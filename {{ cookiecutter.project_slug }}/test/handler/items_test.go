package handler_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"{{ cookiecutter.import_path }}/internal/handler"
	"{{ cookiecutter.import_path }}/internal/middleware"
	"{{ cookiecutter.import_path }}/internal/model"
	"{{ cookiecutter.import_path }}/internal/service"
	"{{ cookiecutter.import_path }}/test/util"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (huma.API, *chi.Mux) {
	t.Helper()

	// Create in-memory SQLite database
	db := util.TestDB(t)
	util.CreateTables(t, db, &model.Item{})

	// Initialize service
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	service.Init(db, logger, "test-secret", "test-api-key")
	t.Cleanup(func() { _ = service.Close() })

	// Create router and API
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("Test", "1.0.0"))

	handler.Setup(api, logger)

	return api, r
}

func TestListItemsEmpty(t *testing.T) {

	_, r := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateAndListItems(t *testing.T) {

	_, r := setupTest(t)

	// Create an item with proper JSON body
	body := strings.NewReader(`{"name":"Test Item","description":"A test item"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// List items
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetItemNotFound(t *testing.T) {

	_, r := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteItemNotFound(t *testing.T) {

	_, r := setupTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateItemNotFound(t *testing.T) {

	_, r := setupTest(t)

	body := strings.NewReader(`{"name":"Updated","description":null}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+uuid.New().String(), body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthFlow(t *testing.T) {

	_, r := setupTest(t)

	// 1. Login to get token
	body := strings.NewReader(`{"api_key":"test-api-key"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var loginResp struct{ Token string }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.Token)

	// 2. Use token to access protected endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func setupTestWithAuth(t *testing.T) (huma.API, *chi.Mux) {
	t.Helper()

	// Create in-memory SQLite database
	db := util.TestDB(t)
	util.CreateTables(t, db, &model.Item{})

	// Initialize service
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	service.Init(db, logger, "test-secret", "test-api-key")
	t.Cleanup(func() { _ = service.Close() })

	// Create router with auth middleware
	r := chi.NewRouter()
	r.Use(middleware.Auth(middleware.AuthOpts{Secret: "test-secret"}))
	api := humachi.New(r, huma.DefaultConfig("Test", "1.0.0"))

	handler.Setup(api, logger)

	return api, r
}

func TestAuthRejectsUnauthorized(t *testing.T) {

	_, r := setupTestWithAuth(t)

	// Request without token should be rejected by auth middleware
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader(`{"name":"Test","description":"Test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthRejectsInvalidToken(t *testing.T) {

	_, r := setupTestWithAuth(t)

	// Request with invalid token should be rejected
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuthAllowsValidToken(t *testing.T) {

	_, r := setupTestWithAuth(t)

	// First get a token via the login endpoint (which is public)
	body := strings.NewReader(`{"api_key":"test-api-key"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp struct{ Token string }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))

	// Now use the token on a protected endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
