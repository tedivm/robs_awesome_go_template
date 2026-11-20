---
name: running-docker-services
description: "Use when starting, stopping, or troubleshooting Docker Compose services for local development. Don't use for production deployment or CI/CD configuration."
license: MIT
compatibility: "Docker, Docker Compose"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Start the development environment with `just docker-up` (or `docker compose up -d`). Services include the app (hot-reload), PostgreSQL, and Redis. Stop with `just docker-down`.

## Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `app` | Built from `Dockerfile` | 8080 | Application server with hot-reload |
| `db` | `postgres:17` | 5432 | PostgreSQL database |
| `redis` | `redis:7-alpine` | 6379 | Redis for caching and task queues |

## Workflow

### Starting the development environment

```bash
# Start all services in detached mode
just docker-up
# Or: docker compose up -d

# Start with logs visible (foreground)
docker compose up

# Rebuild and restart after code changes
docker compose up --build
```

### Development workflow

1. **Start services:** `just docker-up`
2. **Migrations run automatically** via `scripts/dev.sh` (waits for DB, runs goose, starts app)
3. **Code changes hot-reload** via the dev script
4. **View logs:** `docker compose logs -f app`
5. **Stop:** `just docker-down`

### Stopping services

```bash
# Stop services (preserves database volume)
just docker-down
# Or: docker compose down

# Stop and remove all data (including database)
docker compose down -v
```

### Viewing logs

```bash
# Follow app logs
docker compose logs -f app

# Follow database logs
docker compose logs -f db

# View recent logs
docker compose logs --tail=50 app
```

### Adding a new service

Add a new service block to `docker-compose.yaml`:

```yaml
services:
  # ... existing services ...

  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"   # SMTP port
      - "8025:8025"   # Web UI
```

Then restart: `docker compose up -d mailhog`.

### Volume mounts

The `app` service mounts the project directory for hot-reload:

```yaml
volumes:
  - .:/src
```

Database data persists in a named volume:

```yaml
volumes:
  pgdata:
```

To reset the database: `docker compose down -v && docker compose up -d`.

### Environment variables

Services receive environment variables from the `docker-compose.yaml` file and `.env` file (if present):

```yaml
services:
  app:
    environment:
      - APP_SERVER_ENVIRONMENT=development
      - APP_LOG_LEVEL=debug
      - APP_DATABASE_DRIVER=postgres
      - APP_DATABASE_DSN=host=db user=postgres password=postgres dbname=app port=5432 sslmode=disable
```

Override specific variables: `APP_LOG_LEVEL=trace docker compose up app`.

### Dockerfile

Multi-stage build with distroless runtime:

```dockerfile
# ---- Builder ----
FROM golang:{{ cookiecutter.go_version }} AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X {{ cookiecutter.import_path }}/internal/version.Version=${VERSION} -X {{ cookiecutter.import_path }}/internal/version.Commit=${COMMIT}" \
    -o /app/bin/{{ cookiecutter.project_slug }} \
    ./cmd/{{ cookiecutter.project_slug }}

# ---- Runtime ----
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/bin/{{ cookiecutter.project_slug }} ./{{ cookiecutter.project_slug }}

EXPOSE 8080
ENTRYPOINT ["./{{ cookiecutter.project_slug }}"]
```

## Key conventions

- Use `just docker-up/down` — wraps docker compose with consistent behavior
- Database persists in a named volume (`pgdata`) — `docker compose down` preserves data
- Hot-reload via volume mount (`.` to `/src`) — code changes reflect immediately
- `.env` file for local overrides — never commit `.env`, use `.env.example`
- Named volumes for data — `pgdata` for PostgreSQL, add more as needed
- Distroless runtime — no shell, ~2MB image, CGO_ENABLED=0
