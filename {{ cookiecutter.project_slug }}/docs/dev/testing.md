# Testing

This project uses Go's standard `testing` package with [testify](https://github.com/stretchr/testify) for assertions. Tests live in `test/`, mirroring the `internal/` structure.

## Test Organization

```text
test/
├── config/
│   └── config_test.go   # Configuration loading and validation
{% if cookiecutter.include_orm == "y" %}├── service/
│   └── ...              # Service tests (mirrors internal/service/)
{% endif %}{% if cookiecutter.include_web == "y" %}├── handler/
│   └── ...              # Handler tests (mirrors internal/handler/)
├── middleware/
│   └── ...              # Middleware tests (auth, CORS)
{% if cookiecutter.include_orm == "y" %}├── util/
│   └── testdb.go        # Test database helper (SQLite / PostgreSQL)
{% endif %}{% endif %}{% if cookiecutter.include_workers == "y" %}├── worker/
│   └── ...              # Worker pool tests
{% endif %}```

Each `internal/` package has a corresponding `test/` package. This avoids import cycles while keeping tests close to the code they test.

## Running Tests

```bash
# All tests
just test

# Verbose output
just test-verbose

# With coverage
just test-coverage
```

## Test Patterns

### Config Tests

Configuration tests are standalone — they don't need {% if cookiecutter.include_orm == "y" %}a database{% endif %} or external services.

```go
func TestLoadDefaults(t *testing.T) {
    cfg, err := config.Load()
    require.NoError(t, err)
    assert.Equal(t, "info", cfg.Log.Level)
}
```{% if cookiecutter.include_orm == "y" %}
### Service Tests

The service package uses package-level variables (`db`, `logger`) initialized via `service.Init()`. Tests share this state — don't use `t.Parallel()` for service package tests.

```go
func setupTestDB(t *testing.T) *bun.DB {
    t.Helper()
    db := util.TestDB(t)
    _, err := db.NewCreateTable().Model(&model.Item{}).Exec(context.Background())
    require.NoError(t, err)
    service.Init(db, logger, "test-secret", "test-api-key")
    t.Cleanup(func() { _ = service.Close() })
    return db
}
```

The `util.TestDB()` helper creates an in-memory SQLite database by default. Set `APP_DATABASE_DRIVER=postgres` and `APP_DATABASE_DSN` to test against PostgreSQL.{% endif %}{% if cookiecutter.include_web == "y" %}
### Handler Tests

{% if cookiecutter.include_orm == "y" %}For handler tests, initialize the service package with a test database:

```go
func TestGetItem(t *testing.T) {
    db := setupTestDB(t)
    router := chi.NewMux()
    api := humachi.New(router, huma.DefaultConfig("Test", "1.0.0"))

    // Register handler — passes logger, not service
    handler.Setup(api, logger)

    // Create a test item first
    item, _ := service.CreateItem(context.Background(), "Test", "Desc")

    req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+item.ID.String(), nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.StatusCode)
}
```{% else %}For handler tests, use `humatest` or set up a Chi router directly:

```go
func TestHealthCheck(t *testing.T) {
    router := chi.NewMux()
    api := humachi.New(router, huma.DefaultConfig("Test", "1.0.0"))

    handler.Setup(api, logger)

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.StatusCode)
}
```{% endif %}{% endif %}
### Version Tests

Version info is injected at build time via ldflags. The version package exports constants that can be tested directly.

## Coverage

Run `just test-coverage` to generate `coverage.out`. Review it with:

```bash
go tool cover -func=coverage.out
```

Aim for high coverage on the `internal/` packages. The `test/` packages themselves don't need coverage.

## Integration Tests{% if cookiecutter.include_orm == "y" and cookiecutter.include_docker == "y" %}

For PostgreSQL integration tests:

```bash
just test-postgres
```

This spins up a PostgreSQL container, runs the service and handler tests against it, then tears down.{% endif %}
## CI

The test workflow runs on every push and pull request. It checks:

- All tests pass with race detection
- Build succeeds
- No new lint issues