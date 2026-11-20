# {{ cookiecutter.project_name }}

{{ cookiecutter.project_description }}

[![Go Version](https://img.shields.io/badge/Go-{{ cookiecutter.go_version }}-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Lint](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/lint.yaml/badge.svg)](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/lint.yaml)
[![Test](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/test.yaml/badge.svg)](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/test.yaml)
[![Build](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/build.yaml/badge.svg)](https://github.com/{{ cookiecutter.github_org }}/{{ cookiecutter.project_slug }}/actions/workflows/build.yaml)

## Tech Stack

- **Language:** Go {{ cookiecutter.go_version }}
{% if cookiecutter.include_web == "y" %}- **Web Framework:** [Huma v2](https://github.com/danielgtaylor/huma) + [Chi](https://github.com/go-chi/chi)
{% endif %}{% if cookiecutter.include_orm == "y" %}- **ORM:** [bun](https://github.com/uptrace/bun)
- **Migrations:** [goose](https://github.com/pressly/goose)
{% endif %}{% if cookiecutter.include_cli == "y" %}- **CLI:** [cobra](https://github.com/spf13/cobra)
{% endif %}{% if cookiecutter.include_workers == "y" %}- **Workers:** [Asynq](https://github.com/hibiken/asynq)
{% endif %}
- **Config:** [Koanf](https://github.com/knadh/koanf)
- **Logging:** `log/slog` (stdlib)
- **Testing:** testify + gotestsum

## Quick Start

### Prerequisites

- Go {{ cookiecutter.go_version }}+
- [Just](https://justfile.com/) (task runner){% if cookiecutter.include_docker == "y" %}
- Docker + Docker Compose{% endif %}
### Setup

```bash
just install
```

### Pre-commit Hooks (Optional)

Install [pre-commit](https://pre-commit.com/) and set up the hooks:

```bash
pre-commit install
```

This runs `golangci-lint`, `gofmt`, `go vet`, and `go mod tidy` on every commit.{% if cookiecutter.include_web == "y" %}

### Run

```bash
just run
```{% endif %}{% if cookiecutter.include_cli == "y" %}
### CLI

```bash
just run-cli
```{% endif %}{% if cookiecutter.include_docker == "y" %}
### Docker

```bash
just docker-up
```{% endif %}
### Test

```bash
just test
```{% if cookiecutter.include_web == "y" %}

## API

The server exposes a REST API with automatic OpenAPI documentation.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/items` | List items (paginated) |
| `POST` | `/api/v1/items` | Create a new item |
| `GET` | `/api/v1/items/{id}` | Get item by ID |
| `PUT` | `/api/v1/items/{id}` | Update an item |
| `DELETE` | `/api/v1/items/{id}` | Delete an item |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/openapi.json` | OpenAPI 3.0 spec |

### Authentication

Set `APP_AUTH__JWT_SECRET` to enable JWT authentication. When enabled, all requests require a valid Bearer token in the `Authorization` header.

### Pagination

List endpoints support `limit` (1-100, default 20) and `offset` (default 0) query parameters.{% endif %}
## Configuration

Copy `.env.example` to `.env` and adjust settings:

```bash
cp .env.example .env
```

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
{%- if cookiecutter.include_web == "y" %}
| `APP_SERVER__PORT` | `8080` | HTTP port |
{%- endif %}{%- if cookiecutter.include_orm == "y" %}
| `APP_DATABASE__DRIVER` | `sqlite` | Database driver (`sqlite` or `postgres`) |
| `APP_DATABASE__DSN` | — | Database connection string |
{%- endif %}{%- if cookiecutter.include_workers == "y" %}
| `APP_REDIS__ADDR` | `localhost:6379` | Redis address (for workers) |
{%- endif %}
| `APP_LOG__LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
{%- if cookiecutter.include_web == "y" %}
| `APP_AUTH__JWT_SECRET` | — | JWT secret (empty = auth disabled) |
{%- endif %}

## Documentation

See [docs/dev/README.md](docs/dev/README.md) for full documentation.

## License

{% if cookiecutter.license == "None" %}All Rights Reserved
{% else %}[{{ cookiecutter.license }}](LICENSE){% endif %}