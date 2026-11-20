# Justfile — Task Reference

This document catalogs all [just](https://justfile.com/) targets available in the project. Run `just` without arguments to see the list.

## Setup

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `install` | Download Go module dependencies and install required tools{% if cookiecutter.include_orm == "y" %} (goose){% endif %}{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %} (golangci-lint){% endif %} (gotestsum) | Installs binaries to `$GOPATH/bin` |

**When to use:** Run once after cloning the repository, or after updating `go.mod`.

```bash
just install
```

**Idempotent:** Re-running is safe — it re-downloads dependencies and refreshes tool versions.{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}

## Build

| Target | Description | Side Effects |
|--------|-------------|-------------|
{%- if cookiecutter.include_web == "y" %}
| `build` | Compile the application binary to `bin/{{ cookiecutter.project_slug }}` | Creates/overwrites binary in `bin/` |
{%- endif %}{%- if cookiecutter.include_cli == "y" %}
| `build-cli` | Compile the CLI binary to `bin/cli` | Creates/overwrites binary in `bin/` |
{%- endif %}

**When to use:** Before running the binary locally, or to verify the project compiles.

```bash
{% if cookiecutter.include_web == "y" %}just build
{% endif %}{% if cookiecutter.include_cli == "y" %}just build-cli
{% endif %}```

**Idempotent:** Re-running produces a fresh binary with updated version info from git.{% endif %}{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}

## Run

| Target | Description | Side Effects |
|--------|-------------|-------------|
{%- if cookiecutter.include_web == "y" %}
| `dev` | Run with hot-reload via [air](https://github.com/cosmtrek/air) | Starts a long-running process |
| `run` | Run the web server with `go run` | Starts a long-running process |
{%- endif %}{%- if cookiecutter.include_cli == "y" %}
| `run-cli` | Run the CLI tool with `go run` | Starts a CLI command |
{%- endif %}

**When to use:** During development. Use `dev` for hot-reload (recommended), `run` for a simple start, `run-cli` to test CLI commands.

```bash
{% if cookiecutter.include_web == "y" %}just dev
just run
{% endif %}{% if cookiecutter.include_cli == "y" %}just run-cli migrate
{% endif %}```

**Not idempotent:** These are long-running processes. Ctrl+C to stop.{% endif %}
## Test

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `test` | Run all tests with race detector | None |
| `test-verbose` | Run all tests with verbose output and race detector | None |
| `test-coverage` | Run tests with coverage profile and print per-file stats | Creates `coverage.out` |

**When to use:**

- `test` — Quick check before committing. Fails fast on first error.
- `test-verbose` — Debugging a specific test failure. Shows all test names and output.
- `test-coverage` — Checking coverage before a PR. Produces `coverage.out` for review.

```bash
just test
just test-verbose
just test-coverage
```

**Idempotent:** All test targets are safe to re-run.{% if cookiecutter.include_orm == "y" and cookiecutter.include_docker == "y" %}

### PostgreSQL Integration Tests

```bash
just test-postgres
```

Spins up a PostgreSQL container via Docker Compose, runs the service and handler tests against it, then tears down.{% endif %}
## Lint

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `lint` | Run golangci-lint on all packages | None |
| `lint-check` | Run golangci-lint checking only new issues (for CI/pre-commit) | None |

**When to use:**

- `lint` — Full lint check. Use before committing to catch all issues.
- `lint-check` — Check only new issues since last commit. Used by pre-commit hooks and CI.

```bash
just lint
just lint-check
```

**Idempotent:** Both targets are read-only.

## Format

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `fmt` | Format all Go files with `gofmt -w -s` | Modifies files in place |
| `fmt-check` | Check if all Go files are properly formatted | None |

**When to use:**

- `fmt` — Before committing to ensure consistent formatting.
- `fmt-check` — In CI or pre-commit to verify formatting without modifying files.

```bash
just fmt
just fmt-check
```

**Idempotent:** `fmt` modifies files; `fmt-check` is read-only.

## Vet

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `vet` | Run `go vet` on all packages | None |

**When to use:** Catch common errors like incorrect printf formats, unreachable code, etc.

```bash
just vet
```

**Idempotent:** Read-only.{% if cookiecutter.include_orm == "y" %}

## Migrations

| Target | Description | Side Effects |
|--------|-------------|-------------|
{%- if cookiecutter.include_cli == "y" %}
| `migrate-plan` | Show migration status via CLI | None |
| `migrate-apply` | Run pending migrations via CLI | Modifies database schema |
{% else %}
| `migrate-plan` | Show migration status via goose | None |
| `migrate-apply` | Run pending migrations via goose | Modifies database schema |
{% endif %}
| `migrate-create name` | Create new migration files in pg/ and sqlite/ | Creates new files |

**When to use:** After changing the database schema. Create a migration, review it, then apply it.

```bash
{% if cookiecutter.include_cli == "y" %}just migrate-plan
just migrate-apply
{% else %}just migrate-plan
just migrate-apply
{% endif %}just migrate-create add_users_table
```

**Not idempotent:** Migration application modifies the database. Review SQL before applying.{% endif %}{% if cookiecutter.include_docker == "y" %}

## Docker

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `docker-build` | Build the Docker image | Creates a local Docker image |
| `docker-up` | Start the full stack with Docker Compose | Starts containers (db, redis, app) |
| `docker-down` | Stop and remove containers | Stops containers |

**When to use:**

- `docker-build` — Build a production-like image for testing.
- `docker-up` — Start the full development environment with database and cache.
- `docker-down` — Clean up after development.

```bash
just docker-build
just docker-up
just docker-down
```

**Not idempotent:** `docker-up` starts long-running containers. Use `docker-down` to stop them.{% endif %}
## Security

| Target | Description | Side Effects |
|--------|-------------|-------------|
| `vuln-check` | Run [govulncheck](https://go.dev/browse/vuln/) against known vulnerabilities | None |

**When to use:** Before releasing or merging to main. Checks your code against the [Go vulnerability database](https://vuln.go.dev/).

```bash
just vuln-check
```

**Idempotent:** Read-only.

## Chores

| Target | Description | Depends On |
|--------|-------------|-----------|
| `chores` | Run `fmt`, `lint`, and `vet` in sequence | fmt, lint, vet |
| `all` | Full pipeline: install, build, test, lint | install, build{% if cookiecutter.include_cli == "y" %}, build-cli{% endif %}, test, lint |

**When to use:**

- `chores` — Quick quality check before committing.
- `all` — Full CI-like check before creating a PR.

```bash
just chores
just all
```