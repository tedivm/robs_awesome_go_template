# Agent Instructions

You must always follow the best practices outlined in this document. If there is a valid reason why you cannot follow one of these practices, you must inform the user and document the reasons.

Before beginning any task, review the documentation (`docs/dev/` and `README.md`), existing tests, and the `justfile` to understand available dev tools. Review code related to your request to understand preferred style:{% if cookiecutter.include_web == "y" %} review other handlers before creating a new one{% endif %}{% if cookiecutter.include_orm == "y" %}, review existing models before adding a new model{% endif %}, review existing tests before writing new tests.

## Project Overview

{{ cookiecutter.project_name }} is a Go project built with modern, idiomatic conventions.{% if cookiecutter.include_web == "y" %} It uses a layered architecture: handlers (HTTP) → services (business logic) → database.{% endif %}{% if cookiecutter.include_orm == "y" and not cookiecutter.include_web == "y" %} It uses a layered architecture: services (business logic) → database.{% endif %}

### Tech Stack

| Area | Technology | Purpose |
|------|-----------|---------|
{% if cookiecutter.include_web == "y" %}| Router | [Chi](https://github.com/go-chi/chi) v5 | HTTP routing and middleware |
| API | [Huma](https://github.com/danielgtaylor/huma) v2 | OpenAPI-first API with auto-validation |
{% endif %}{% if cookiecutter.include_orm == "y" %}| ORM | [bun](https://github.com/uptrace/bun) v1 | SQL-first ORM |
| Migrations | [goose](https://github.com/pressly/goose) v3 | Versioned SQL migrations |
{% endif %}| Config | [Koanf](https://github.com/knadh/koanf) | YAML + env var configuration |
{% if cookiecutter.include_cli == "y" %}| CLI | [cobra](https://github.com/spf13/cobra) | Command-line interface |
{% endif %}{% if cookiecutter.include_workers == "y" %}| Workers | [asynq](https://github.com/hibiken/asynq) + Go channels | Distributed and local background tasks |
{% endif %}| Testing | [testify](https://github.com/stretchr/testify) + stdlib `testing` | Table-driven tests, assertions |
| Logging | `log/slog` (stdlib) | Structured logging |
| Linting | [golangci-lint](https://github.com/golangci/golangci-lint) | Multi-linter with caching |
| Task Runner | [just](https://justfile.com/) | Dev task orchestration |
{% if cookiecutter.include_docker == "y" %}| Docker | Multi-stage build + distroless runtime | Production container images |
{% endif %}

### Project Layout

```
{% if cookiecutter.include_web == "y" %}cmd/{{ cookiecutter.project_slug }}/         Web server entry point
{% endif %}{% if cookiecutter.include_cli == "y" %}cmd/cli/                                     CLI entry point
{% endif %}internal/
  config/                                    Configuration loading
{% if cookiecutter.include_orm == "y" %}  model/                                     Domain models (bun)
{% endif %}{% if cookiecutter.include_web == "y" %}  handler/                                   HTTP handlers (Huma)
{% endif %}{% if cookiecutter.include_orm == "y" %}  service/                                   Business logic
{% endif %}{% if cookiecutter.include_workers == "y" %}  worker/                                    Background task definitions
{% endif %}{% if cookiecutter.include_cli == "y" %}  cli/                                       Cobra commands
{% endif %}  version/                                   Build-time version info
{% if cookiecutter.include_orm == "y" %}migrations/                                  Goose SQL migrations
{% endif %}configs/                                     YAML configuration files
{% if cookiecutter.include_web == "y" %}static/                                      Static assets
{% endif %}test/                                        Test suites (mirrors internal/)
docs/dev/                                    Developer documentation
{% if cookiecutter.include_docker == "y" %}scripts/                                     Dev scripts (e.g., dev.sh)
{% endif %}```

## Important Commands

### Setup

```bash
just install       # Install dependencies and tools
```

{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}
### Build and Run

```bash
{% if cookiecutter.include_web == "y" %}just build         # Build binary with version info
{% endif %}{% if cookiecutter.include_cli == "y" %}just build-cli     # Build CLI binary
{% endif %}{% if cookiecutter.include_web == "y" %}just run           # Run the web server locally
{% endif %}{% if cookiecutter.include_cli == "y" %}just run-cli       # Run the CLI tool
{% endif %}```
{% endif %}

### Testing

```bash
just test          # Run all tests with race detector
just test-verbose  # Run tests with verbose output
just test-coverage # Run tests with coverage report
```

### Code Quality

```bash
just lint          # Run golangci-lint
just lint-check    # Check for new lint issues only (pre-commit)
just fmt           # Format all Go files
just fmt-check     # Check formatting without changes
just vet           # Run go vet
just chores        # Run fmt, lint, vet together
```

{% if cookiecutter.include_orm == "y" %}
### Database

```bash
just migrate-plan    # Show migration status
just migrate-apply   # Apply all pending migrations
just migrate-create  # Create a new migration file (e.g., just migrate-create add_pages)
```
{% endif %}

{% if cookiecutter.include_docker == "y" %}
### Docker

```bash
just docker-build    # Build the Docker image
just docker-up       # Start all services with docker compose
just docker-down     # Stop all services
```
{% endif %}

### Security

```bash
just vuln-check      # Run govulncheck against codebase
```

## Best Practices

### Testing (Required)

**Every new feature MUST include passing tests before it is considered complete.**

- Tests must cover the happy path, error cases, and edge cases
- Run `just test` after every change — do not commit if tests fail
- Use table-driven tests with `t.Run()` subtests for all test functions
- Use `require` for setup (stops on failure), `assert` for checks (continues)
{% if cookiecutter.include_orm == "y" %}- Use in-memory SQLite for database tests — no external dependencies
{% endif %}{% if cookiecutter.include_web == "y" %}- Use `humatest` for API handler tests — in-process, no network needed
{% endif %}- Target 80%+ line coverage
- Run with `-race` flag to catch data races
- Test packages use `_test` suffix: `package{% if cookiecutter.include_web == "y" %} handler_test{% elif cookiecutter.include_orm == "y" %} service_test{% else %} config_test{% endif %}`

### General

- Write idiomatic Go. Follow the [Uber Go Style Guide](https://github.com/uber-go/guide) and [Effective Go](https://go.dev/doc/effective_go).
- Use `{{ cookiecutter.import_path }}` as the module import path.
- Place all application code in `internal/` — it cannot be imported by external packages.
- Entry points live in `cmd/`. Each binary gets its own subdirectory.
- Prefer composition over inheritance. Embed structs for shared behavior.
- Use interfaces sparingly. Define them where they are consumed, not where they are implemented.
- Do not put `import` statements inside functions unless necessary to prevent circular imports.

{% if cookiecutter.include_orm == "y" or cookiecutter.include_web == "y" %}
### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Check sentinel errors with `errors.Is(err, sentinel)`
- Unwrap typed errors with `errors.As(err, &target)`
- Join multiple errors with `errors.Join(err1, err2)` (Go 1.20+)
- Use `errgroup` for concurrent operations with error handling
- Never swallow errors silently. Log or return them.

```go
// Good: Contextual error wrapping
func GetItem(ctx context.Context, id uuid.UUID) (ItemRead, error) {
    item := &model.Item{}
    err := db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ItemRead{}, fmt.Errorf("get item %s: %w", id, ErrNotFound)
        }
        return ItemRead{}, fmt.Errorf("get item %s: %w", id, err)
    }
    return toItemRead(*item), nil
}
```
{% endif %}

### Logging

- Use `log/slog` (stdlib) for all logging. Never use `fmt.Println` or `print`.
- Create a logger in `main.go` and pass it through constructors.
- Use `logger.With("key", value)` to add context to loggers.
- Use appropriate levels: `Debug` for tracing, `Info` for events, `Warn` for issues, `Error` for failures.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
```

{% if cookiecutter.include_web == "y" %}
### Security

- Never hardcode sensitive data. Use environment variables or config files.
- Do not log sensitive data (passwords, tokens, PII).
- All user input must be validated via Huma struct tags.
- Use UUIDs for resource identifiers — never expose sequential IDs.
{% endif %}

### Production Ready

- All generated code is production-ready. No stubs or dev-only branches in `internal/`.
- Dev/production differences are configuration, not code.
- Run `just chores` before committing to ensure formatting and linting pass.

## Files

- Filenames must be lowercase (case-insensitive filesystem compatibility).
- Developer documentation lives in `docs/dev/`.
- New developer documents must be added to `docs/dev/README.md`.
{% if cookiecutter.include_orm == "y" %}- Database models live in `internal/model/`.
{% endif %}- Configuration lives in `internal/config/config.go` and `configs/config.yaml`.
- `.env` is in `.gitignore`. Use `.env.example` as a template.
- Use `git mv` for moving or renaming tracked files.
