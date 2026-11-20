# Rob's Awesome Go Template

**A Go project template for building production-ready applications with modern tooling**

This Cookiecutter template generates complete Go projects with idiomatic structure, modern frameworks, and production-ready infrastructure. Features are independently selectable — you get what you need, nothing else.

**Quick Links:** [Quick Start](#quick-start) • [Core Features](#core-features-always-included) • [Optional Integrations](#optional-integrations)

## What It Generates

- **Go-Native Layout**: `cmd/` + `internal/` structure with flat packages, following [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- **Modern Tooling**: Huma v2 for OpenAPI-first APIs, bun for SQL-first ORM, goose for migrations, Koanf for config, golangci-lint for linting, Justfile for task automation
- **Production Infrastructure**: Multi-stage Docker builds with distroless runtime, independent GitHub Actions workflows, Go Releaser with cosign signing, SBOM generation with Syft
- **Modular Components**: Features are independently selectable. Choose what you need — web API, ORM, CLI, workers, Docker, CI — and the template generates only what you asked for
- **AI Ready**: Pre-configured AGENTS.md file following the [open standard](https://agents.md/) compatible with Cursor, Aider, GitHub Copilot, and other AI coding assistants. Per-feature agent skills in `.agents/skills/` provide framework-specific guidance for Huma routes, bun models, and other components
- **Go 1.27+ Standards**: `log/slog` structured logging, table-driven tests, proper module layout

## Quick Start

```bash
# Create your project — uv handles installing cookiecutter automatically
uv tool run cookiecutter gh:tedivm/robs_awesome_go_template
```

Answer a few questions about your project, and you'll have a complete Go application with everything configured and ready to go. Requires [uv](https://github.com/astral-sh/uv).

## Core Features (Always Included)

Every project created with this template includes these essential components:

### Project Structure

- **Idiomatic layout**: `cmd/` for entry points, `internal/` for private packages — following [golang-standards/project-layout](https://github.com/golang-standards/project-layout) and [official Go module layout](https://go.dev/doc/modules/layout)
- **Flat package organization**: No `pkg/` directory. Packages live directly under `internal/` with clear boundaries
- **Justfile automation**: Modern task runner covering setup, testing, formatting, linting, and deployment with simple commands like `just test` and `just build`

### Configuration

- **[Koanf](https://github.com/knadh/koanf) configuration**: Modular, case-sensitive config library with YAML file and environment variable providers
- **Hierarchical settings**: `configs/config.yaml` for defaults, environment variables (with `APP_` prefix) for overrides
- **Automatic validation**: Runtime defaults for missing values, clear error messages for misconfigurations

### Code Quality

- **[golangci-lint](https://github.com/golangci/golangci-lint)**: Runs 100+ linters in parallel with caching, configured via `.golangci.yml`
- **[Pre-commit hooks](https://pre-commit.com/)**: Automated code quality checks before every commit — linting, formatting, vet
- **Auto-formatting**: `gofmt` with simplify flag, enforced via pre-commit and CI

### Testing

- **[testify](https://github.com/stretchr/testify)**: Industry-standard assertions with `assert` and `require` packages
- **Table-driven tests**: Go-native test patterns with `t.Run()` subtests
- **In-memory test database**: SQLite for fast, dependency-free database testing via `modernc.org/sqlite`
- **Coverage reporting**: `go test -coverprofile` with `go tool cover` for function-level coverage output

### Logging

- **[log/slog](https://pkg.go.dev/log/slog)**: Stdlib structured logging with JSON and text handlers, leveled logging, context fields
- **Per-package loggers**: Module-tagged loggers for traceable output across services
- **Environment-aware**: Development gets text format with source locations; production gets JSON

### Versioning

- **Dynamic version injection**: Build-time version, commit hash, and build date via `ldflags`
- **Go Releaser**: Automated releases with multi-platform binaries, checksums, and changelog generation
- **Conventional Commits**: Structured commit messages for automatic changelog generation

### Documentation

- **Developer Guide Hub**: Organized documentation in `docs/dev/` with dedicated guides for each enabled feature
- **CHANGELOG.md**: Keep a Changelog format, updated per release

## Optional Integrations

Pick and choose the features you need. Unused components are completely removed from your project.

### Web APIs

**[Huma v2](https://github.com/danielgtaylor/huma) + [Chi](https://github.com/go-chi/chi)**

- OpenAPI-first API framework — automatic OpenAPI 3.1 documentation, request validation, and RFC 9457 error responses, all derived from Go struct tags
- Struct-based endpoints: define inputs and outputs as Go structs, Huma generates the OpenAPI spec and validates requests automatically
- Tag-based validation: `maxLength`, `minimum`, `maximum`, `pattern`, `enum` — all from struct tags, no separate validation library needed
- Interactive API docs served automatically at `/docs` with Swagger UI, plus `/openapi.json` and `/openapi.yaml` endpoints
- Chi router with battle-tested middleware: recovery, request logging, CORS, rate limiting, body limits, compression
- Built-in CLI support via `humacli` — automatic `--port`, `-p`, and `SERVICE_PORT` env var handling
- HTML templating with Go's stdlib `html/template` for server-rendered pages alongside the API
- Static file serving with optimized caching headers via Chi

### Database & ORM

**[bun](https://github.com/uptrace/bun) + [goose](https://github.com/pressly/goose)**

- bun is a "SQL-first" ORM — embraces SQL rather than hiding it. Type-safe query building, flexible scanning, multi-database support (PostgreSQL, MySQL, SQLite, MSSQL)
- Base model with UUID primary keys, automatic timestamps, and soft deletes
- Cross-database compatibility: SQLite for local development, PostgreSQL for production — same code, different config
- goose-powered versioned SQL migrations for schema changes, with both up and down migrations
- Connection pool management with configurable pool sizes, overflow handling, and health checks

### Command Line Interface

**[cobra](https://github.com/spf13/cobra)**

- Industry-standard CLI framework used by kubectl, Hugo, and most major Go CLI tools
- Subcommands, flags, help text, and auto-generated man pages
- Pre-configured with `migrate` and `version` subcommands
- Seamless integration with the rest of the project's config and service layers

### Background Workers

**Go worker pools + [Asynq](https://github.com/hibiken/asynq)**

- Go worker pools for CPU-bound tasks: standard channels and goroutines, no external dependency
- Asynq for distributed I/O-bound tasks: Redis-backed task queue with retry, scheduling, priority queues, and result tracking
- Pre-configured task types with automatic serialization, retry with exponential backoff, and graceful shutdown
- Separate server and scheduler components for different deployment patterns

### Containerization

**[Docker](https://www.docker.com/) with distroless runtime**

- Multi-stage builds: `golang` builder stage with CGO disabled, `gcr.io/distroless/static` runtime (~2MB, no shell)
- Version injection via build args for clean binary versioning
- Docker Compose for local development with hot-reload, database, and Redis services
- Development script that waits for dependencies, runs migrations, and starts the application
- Multi-architecture builds (AMD64, ARM64) via GitHub Actions

### CI/CD Pipeline

**Comprehensive [GitHub Actions](https://docs.github.com/en/actions)**

- Independent workflows for faster feedback: lint, test, build, Docker, release, security, license compliance
- Matrix testing with race detector and coverage reporting
- Automated releases via Go Releaser v2 with cosign keyless signing and Syft SBOM generation
- Security scanning with `govulncheck` and Trivy container scanning
- Dependabot for automated dependency updates
- License compliance checking with `go-licenses`

### AI Development Support

- Pre-configured **AGENTS.md** file following the [open standard](https://agents.md/) adopted by over 20,000 open-source projects
- Per-feature **agent skills** in `.agents/skills/` providing framework-specific guidance for Huma routes, bun models, cobra commands, and other components
- Comprehensive project-specific guidelines including Go version requirements, slog preferences, typing conventions, and framework-specific patterns
- Detailed best practices for production-ready code including error handling, dependency management, and testing patterns
- Context-aware instructions that adapt based on your selected features, ensuring AI assistants understand your project's architecture

## Smart Configuration

The template intelligently configures itself based on your choices through sophisticated post-generation hooks:

- **Surgical cleanup**: Only includes files and packages you actually need. Disabled features are removed entirely — no dead code, no commented-out sections
- **Conditional Docker services**: Docker Compose generates only the services your project requires: PostgreSQL for ORM, Redis for workers or caching
- **Feature-driven CI/CD**: GitHub Actions workflows are conditionally installed based on your selections: Docker builds only when Docker is enabled, release workflows only when CLI or web is enabled
- **Framework integration**: Automatically wires together selected components (Huma with bun database dependencies, cobra with migration commands) providing working examples of how pieces fit together
- **Dependency management**: `go mod tidy` runs automatically post-generation, ensuring clean, minimal `go.mod` and `go.sum` files

## Multiple License Options

Choose the license that fits your project:

- **MIT**: Permissive, business-friendly
- **Apache-2.0**: Patent protection included
- **BSD-3-Clause**: Simple and permissive
- **GPL-3.0**: Copyleft for open source
- **All Rights Reserved**: Proprietary projects
- **None**: No license file generated

## Acknowledgments

This template was inspired by [Rob's Awesome Python Template](https://github.com/tedivm/robs_awesome_python_template), applying the same philosophy of modular, production-ready project scaffolding to the Go ecosystem.

## Get Started

```bash
cookiecutter gh:tedivm/robs_awesome_go_template
```

Answer the prompts and you'll have a complete Go project with your selected features configured and ready to go.
