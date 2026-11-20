# {{ cookiecutter.project_name }} — Developer Documentation

This directory contains developer-facing documentation for {{ cookiecutter.project_name }}. Each document covers a specific aspect of the project.

## Quick Start

### Setup

```bash
# Install dependencies and tools
just install
```

This downloads Go module dependencies and installs required tooling.{% if cookiecutter.include_orm == "y" %} goose{% endif %}{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %} golangci-lint{% endif %} gotestsum.{% if cookiecutter.include_web == "y" %}

### Run the Server

```bash
# Start the development server
just run
```

The server listens on `http://localhost:8080` by default. Visit `http://localhost:8080/docs` for the interactive API documentation.{% endif %}{% if cookiecutter.include_cli == "y" %}

### Run the CLI

```bash
# Run the CLI
just run-cli
```{% endif %}
### Run Tests

```bash
# Run all tests
just test

# Run with verbose output
just test-verbose

# Run with coverage report
just test-coverage
```

### Make Your First Change

1. Edit a file in `internal/` — this is where all application code lives.
2. Run `just test` to verify your changes don't break anything.
3. Run `just chores` to format, lint, and vet your code.
4. Commit with a [Conventional Commit](https://www.conventionalcommits.org/) message:{% if cookiecutter.include_web == "y" %} `feat(handler): add pagination support`{% elif cookiecutter.include_cli == "y" %} `feat(cli): add migrate command`{% else %} `feat(config): add new setting`{% endif %}

## Project Overview

{{ cookiecutter.project_name }} is a Go project generated from [Rob's Awesome Go Template](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}). It follows idiomatic Go conventions with a modular, production-first architecture.

### Architecture

```mermaid
graph LR
{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}  subgraph cmd
{% if cookiecutter.include_web == "y" %}    A["cmd/{{ cookiecutter.project_slug }}/main.go"]
{% endif %}{% if cookiecutter.include_cli == "y" %}    B["cmd/cli/main.go"]
{% endif %}  end
{% endif %}  subgraph internal
    C["config"]
{% if cookiecutter.include_orm == "y" %}    D["model"]
{% endif %}{% if cookiecutter.include_web == "y" %}    E["handler"]
{% endif %}{% if cookiecutter.include_orm == "y" %}    F["service"]
{% endif %}{% if cookiecutter.include_workers == "y" %}    H["worker"]
{% endif %}{% if cookiecutter.include_cli == "y" %}    I["cli"]
{% endif %}    J["version"]
  end
{% if cookiecutter.include_orm == "y" or cookiecutter.include_workers == "y" %}
  subgraph external
{% if cookiecutter.include_orm == "y" %}    K["Database<br/>(SQLite / PostgreSQL)"]
{% endif %}{% if cookiecutter.include_workers == "y" %}    L["Redis<br/>(Asynq)"]
{% endif %}  end
{% endif %}{% if cookiecutter.include_web == "y" %}
  A --> C
  A --> E
{% if cookiecutter.include_orm == "y" %}  A --> F
{% endif %}{% if cookiecutter.include_cli == "y" %}  B --> I
{% endif %}{% if cookiecutter.include_web == "y" and cookiecutter.include_orm == "y" %}  E --> F
{% endif %}{% if cookiecutter.include_orm == "y" %}  F --> K
{% endif %}{% if cookiecutter.include_orm == "y" and cookiecutter.include_workers == "y" %}  F --> H
{% endif %}{% if cookiecutter.include_workers == "y" %}  H --> L
{% endif %}{% endif %}
```

### Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Entry points — one per binary |
| `internal/` | Private application packages |
{%- if cookiecutter.include_orm == "y" %}
| `migrations/` | Versioned SQL migrations (goose) |
{%- endif %}
| `configs/` | YAML configuration files |
| `test/` | Tests mirroring `internal/` structure |
{%- if cookiecutter.include_docker == "y" %}
| `scripts/` | Development scripts |
{%- endif %}
| `docs/dev/` | This documentation |

### Tech Stack

| Layer | Choice |
|-------|--------|
{%- if cookiecutter.include_web == "y" %}
| Web Framework | [Huma v2](https://github.com/danielgtaylor/huma) + [Chi](https://github.com/go-chi/chi) |
{%- endif %}{%- if cookiecutter.include_orm == "y" %}
| ORM | [bun](https://github.com/uptrace/bun) |
| Migrations | [goose](https://github.com/pressly/goose) |
{%- endif %}{%- if cookiecutter.include_cli == "y" %}
| CLI | [cobra](https://github.com/spf13/cobra) |
{%- endif %}{%- if cookiecutter.include_workers == "y" %}
| Workers | Go channels (CPU) + [Asynq](https://github.com/hibiken/asynq) (I/O) |
{%- endif %}
| Configuration | [Koanf](https://github.com/knadh/koanf) |
| Testing | [testify](https://github.com/stretchr/testify) + stdlib `testing` |
| Linting | [golangci-lint](https://github.com/golangci/golangci-lint) |
{%- if cookiecutter.include_docker == "y" %}
| Docker | Multi-stage build, distroless runtime |
{%- endif %}
| CI/CD | GitHub Actions{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}, Go Releaser{% endif %} |

## Documentation Index

| Document | Audience | Covers |
|----------|----------|--------|
| [justfile.md](justfile.md) | All developers | Task runner targets, when to use each |
{%- if cookiecutter.include_web == "y" %}
| [api.md](api.md) | Developers, API consumers | REST API design, Huma + Chi, OpenAPI |
{%- endif %}{%- if cookiecutter.include_orm == "y" %}
| [database.md](database.md) | Developers | bun ORM, models, queries, connection config |
| [migrations.md](migrations.md) | Developers | goose migration workflow, SQL format |
{%- endif %}
| [testing.md](testing.md) | Developers | Test patterns, organization, coverage |
{%- if cookiecutter.include_docker == "y" %}
| [docker.md](docker.md) | Developers, DevOps | Docker Compose, Dockerfile, deployment |
{%- endif %}
| [github.md](github.md) | Developers, maintainers | CI/CD workflows, releases, pre-commit |
{%- if cookiecutter.include_cli == "y" %}
| [cli.md](cli.md) | Developers | Cobra CLI structure, adding commands |
{%- endif %}{%- if cookiecutter.include_workers == "y" %}
| [workers.md](workers.md) | Developers | Background workers, Asynq, task types |
{%- endif %}
| [configuration.md](configuration.md) | Developers | Config sources, precedence, secrets |
| [security.md](security.md) | Developers, security reviewers | Input validation{% if cookiecutter.include_web == "y" %}, auth, CORS{% endif %}, secrets{% if cookiecutter.include_web == "y" %}, HTTPS{% endif %} |
