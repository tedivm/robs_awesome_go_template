---
name: writing-go-tests
description: "Use when writing, modifying, or debugging Go tests. Covers table-driven tests, API handler tests with humatest, database tests with in-memory SQLite, and mockery. Don't use for linting or code formatting."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+, testify, mockery"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Write table-driven tests with `t.Run()` subtests in `test/` (mirrors `internal/`). Use `require` for setup, `assert` for checks. Use in-memory SQLite for database tests, humatest for API tests.

## Test structure

Tests live in `test/` and mirror the `internal/` package structure:

```
test/
├── util/
│   └── testdb.go          # Shared test utilities
├── handler/
│   └── health_test.go     # Handler tests (mirrors internal/handler/)
├── service/
│   └── ...                # Service tests (mirrors internal/service/)
├── config/
│   └── config_test.go     # Config loading tests
└── integration/
    └── ...                # Full-stack integration tests
```

## Workflow

### Table-driven tests

All tests use this pattern:

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        cfg     *config.Config
        wantErr bool
    }{
        {
            name: "valid config",
            cfg: &config.Config{
                Server: config.ServerConfig{Port: 8080},
            },
            wantErr: false,
        },
        {
            name: "empty port",
            cfg: &config.Config{
                Server: config.ServerConfig{Port: 0},
            },
            wantErr: true,
        },
        {
            name: "invalid environment",
            cfg: &config.Config{
                Server: config.ServerConfig{Environment: "invalid"},
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.cfg.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### API handler tests with humatest

```go
func TestCreateItem(t *testing.T) {
    // Setup — init the service package with a test database
    db := util.TestDB(t)
    _, err := db.NewCreateTable().Model(&model.Item{}).Exec(context.Background())
    require.NoError(t, err)

    logger := slog.Default()
    service.Init(db, logger, "test-secret", "test-api-key")
    t.Cleanup(func() { _ = service.Close() })

    // Create router and API
    r := chi.NewRouter()
    api := humachi.New(r, huma.DefaultConfig("Test", "1.0.0"))
    handler.Setup(api, logger)

    // Test valid request
    body := strings.NewReader(`{"name":"Test Item","description":"A test item"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)

    // Test missing required field
    body = strings.NewReader(`{"description":"Missing name"}`)
    req = httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
```

### Database tests with in-memory SQLite

```go
// test/util/testdb.go
package util

import (
    "database/sql"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
    _ "modernc.org/sqlite"
)

func TestDB(t *testing.T) *bun.DB {
    t.Helper()

    conn, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)

    db := bun.NewDB(conn, sqlitedialect.New())

    t.Cleanup(func() {
        conn.Close()
    })

    return db
}
```

### Service tests

```go
func TestService_CreateItem(t *testing.T) {
    // Setup — init the service package
    db := util.TestDB(t)
    _, err := db.NewCreateTable().Model(&model.Item{}).Exec(context.Background())
    require.NoError(t, err)

    service.Init(db, slog.Default(), "test-secret", "test-api-key")
    t.Cleanup(func() { _ = service.Close() })

    tests := []struct {
        name        string
        nameInput   string
        description string
        wantErr     bool
    }{
        {
            name:        "valid item",
            nameInput:   "Test Item",
            description: "A test item",
            wantErr:     false,
        },
        {
            name:        "empty name",
            nameInput:   "",
            description: "A test item",
            wantErr:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            item, err := service.CreateItem(context.Background(), tt.nameInput, tt.description)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, item.ID)
                assert.Equal(t, tt.nameInput, item.Name)
            }
        })
    }
}
```

### Mocking external dependencies

For dependencies like HTTP clients, cache providers, or message queues, define small interfaces in the consumer's package and mock them:

```go
// In the service package (unexported interface)
type emailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// In tests
func TestService_WithEmail(t *testing.T) {
    mockSender := &mockEmailSender{}
    mockSender.onSend = func(ctx context.Context, to, subject, body string) error {
        return nil
    }
    // ...
}
```

Prefer simple hand-rolled mocks over mockery for small interfaces. Use mockery only for larger, generated interfaces.

## Running tests

```bash
just test           # Run all tests with race detector
just test-verbose   # Verbose output
just test-coverage  # Coverage report
```

Or directly:

```bash
go test -race -count=1 ./...
go test -v -race -count=1 ./internal/service/...
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Gotchas

**`humatest.New()` returns `(http.Handler, error)`, not `(huma.API, error)`.**

Use `humachi.New()` directly for handler tests:

```go
r := chi.NewRouter()
api := humachi.New(r, huma.DefaultConfig("Test", "1.0.0"))
```

**Test table setup: use `util.CreateTables` for concise setup.**

Each test file needs its tables created. Use the `util.CreateTables` helper:

```go
func setupTestDB(t *testing.T) *bun.DB {
    db := util.TestDB(t)
    util.CreateTables(t, db, &model.Category{}, &model.Page{})
    // List tables in dependency order: parents before children
    service.Init(db, slog.Default(), "test-secret", "test-api-key")
    t.Cleanup(func() { _ = service.Close() })
    return db
}
```

**No `t.Parallel()` for tests using the service package.**

Package-level variables (`db`, `logger`, `jwtSecret`) are shared state. Tests that call `service.Init()` cannot run in parallel — they will race on the globals and produce flaky failures.

## Key conventions

- Table-driven tests with subtests — every test function uses this pattern
- `require` for setup, `assert` for checks — `require` stops, `assert` continues
- `t.Helper()` on wrapper functions — correct line numbers in output
- In-memory SQLite for DB tests — fast, no external dependencies
- humachi for API tests — in-process HTTP testing, no network needed
- **No `t.Parallel()` for tests using the service package** — package-level variables are shared state
- Target 80%+ line coverage — tracked in CI with Codecov
- Run with `-race` — catches data races in CI and locally
- Test file names end in `_test.go` — Go convention
- Test packages use `_test` suffix — e.g., `package handler_test`
