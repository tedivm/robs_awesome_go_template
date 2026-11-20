# Docker

{{ cookiecutter.project_name }} uses Docker for local development and production deployment. The project uses a multi-stage Dockerfile with named build targets and separate Docker Compose configurations for development and production.

## Docker Compose Services

### Development (`docker-compose.yaml`)

The development configuration uses the Go compiler image with hot-reload via volume mounts.

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
{% if cookiecutter.include_web == "y" or cookiecutter.include_cli == "y" %}| `app` | Go compiler (builder stage) |{% if cookiecutter.include_web == "y" %} 8080 | Application server with hot-reload{% endif %}{% if cookiecutter.include_cli == "y" and not cookiecutter.include_web == "y" %} - | CLI binary{% endif %} |
{% endif %}{% if cookiecutter.include_orm == "y" %}| `db` | `postgres:17` | 5432 | PostgreSQL database |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `redis` | `redis:7-alpine` | 6379 | Redis for Asynq task queue |
{% endif %}

### Production (`docker-compose.prod.yaml`)

The production configuration uses the minimal distroless runtime images with separate services for independent scaling.

| Service | Target | Port | Purpose |
|---------|--------|------|---------|
{% if cookiecutter.include_web == "y" %}| `web` | `web-build` | 8080 | HTTP API server |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `worker` | `worker-build` | — | Background task processor |
{% endif %}{% if cookiecutter.include_orm == "y" %}| `db` | `postgres:17` | 5432 | PostgreSQL database |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `redis` | `redis:7-alpine` | 6379 | Redis for Asynq task queue |
{% endif %}

### Starting the Stack

```bash
# Development (hot-reload)
just docker-up

# Production (distroless images)
just docker-up-prod
```

### Stopping the Stack

```bash
just docker-down

# Stop production stack
docker compose -f docker-compose.prod.yaml down
```

## Local Development with Hot-Reload

The `app` service in Docker Compose is configured for development:

```yaml
services:
  app:
    build:
      context: .
      target: builder  # Uses Go compiler image for hot-reload
{% if cookiecutter.include_web == "y" %}    ports:
      - "8080:8080"
{% endif %}    environment:
      - APP_SERVER_ENVIRONMENT=development
      - APP_LOG_LEVEL=debug
{% if cookiecutter.include_orm == "y" %}      - APP_DATABASE_DRIVER=postgres
      - APP_DATABASE_DSN=host=db user=postgres password=postgres dbname=app port=5432 sslmode=disable
{% endif %}{% if cookiecutter.include_workers == "y" %}      - APP_REDIS_ADDR=redis:6379
{% endif %}    volumes:
      - .:/src
    working_dir: /src
{% if cookiecutter.include_orm == "y" and cookiecutter.include_cli == "y" %}    command: sh -c "until pg_isready -h db -p 5432 -U postgres; do sleep 1; done && go run ./cmd/cli migrate up &&{% if cookiecutter.include_web == "y" %} go run ./cmd/{{ cookiecutter.project_slug }}{% endif %}"
{% elif cookiecutter.include_web == "y" %}    command: go run ./cmd/{{ cookiecutter.project_slug }}
{% elif cookiecutter.include_cli == "y" %}    command: go run ./cmd/cli
{% endif %}```

The volume mount (`- .:/src`) syncs your local files into the container. Migrations run automatically via the CLI before the app starts. For local hot-reload without Docker, use `just dev` (runs [air](https://github.com/cosmtrek/air)).

### Development Workflow

1. Start the stack: `just docker-up`
2. Edit code in your local editor
3. Save — the container detects the change and restarts
4. Visit `http://localhost:8080` to see changes
5. Check logs: `docker compose logs -f app`

### Database Access

The PostgreSQL database is accessible at `localhost:5432`:

```bash
# Connect with psql
docker compose exec db psql -U postgres -d app

# Or use a database client with these settings:
# Host: localhost
# Port: 5432
# User: postgres
# Password: postgres
# Database: app
```

## Multi-Stage Dockerfile with Named Targets

The `Dockerfile` uses a shared builder stage and separate runtime targets. This produces minimal distroless images for each component.

### Build Targets

| Target | Binary | Description |
|--------|--------|-------------|
{% if cookiecutter.include_web == "y" %}| `web-build` | `{{ cookiecutter.project_slug }}` | HTTP API server |
{% endif %}{% if cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}| `cli-build` | `cli` | CLI tool (migrations, admin) |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `worker-build` | `worker` | Background task processor |
{% endif %}| `builder` | — | Go compiler image (for dev hot-reload) |

### Building Specific Targets

```bash
# Build all targets (default)
just docker-build

# Build a specific target
docker build --target web-build -t {{ cookiecutter.project_slug }}:web .
docker build --target cli-build -t {{ cookiecutter.project_slug }}:cli .
{% if cookiecutter.include_workers == "y" %}docker build --target worker-build -t {{ cookiecutter.project_slug }}:worker .
{% endif %}# Build with version info
docker build \
    --build-arg VERSION=1.0.0 \
    --build-arg COMMIT=$(git rev-parse --short HEAD) \
    --build-arg DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
    --target web-build \
    -t {{ cookiecutter.project_slug }}:1.0.0 .
```

### Builder Stage

```dockerfile
FROM golang:{{ cookiecutter.go_version }} AS builder

WORKDIR /src

# Copy dependency definitions first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
```

Key points:

- **Dependency caching:** `go.mod` and `go.sum` are copied and downloaded before source code. Docker caches this layer — subsequent builds skip `go mod download` unless dependencies change.
- **CGO disabled:** `CGO_ENABLED=0` produces a static binary with no C dependencies.
- **Version injection:** `-ldflags` sets version, commit, and date at build time.

### Runtime Stages

Each target gets its own minimal distroless runtime:

```dockerfile
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=web-build /app/{{ cookiecutter.project_slug }} .
{% if cookiecutter.include_web == "y" %}COPY --from=builder /src/static ./static

EXPOSE 8080{% endif %}

ENTRYPOINT ["./{{ cookiecutter.project_slug }}"]
```

Key points:

- **Distroless base:** `gcr.io/distroless/static-debian12` is a minimal image (~2MB) with no shell, no package manager, no unnecessary binaries. Smaller attack surface.
- **Single binary:** Only the compiled Go binary is in the final image. No Go toolchain, no source code.
- **Direct entrypoint:** Runs the binary directly — no shell wrapper.

## Production Deployment

### Starting the Production Stack

```bash
# Build and start all production services
just docker-up-prod

# Or with explicit compose file
docker compose -f docker-compose.prod.yaml up -d

# Stop the production stack
docker compose -f docker-compose.prod.yaml down
```

### Running Individual Services

```bash
{% if cookiecutter.include_web == "y" %}# Web server only
docker run -d \
    -p 8080:8080 \
    -e APP_SERVER_ENVIRONMENT=production \
{% if cookiecutter.include_orm == "y" %} \
    -e APP_DATABASE_DRIVER=postgres \
    -e APP_DATABASE_DSN="host=db user=postgres password=SECRET dbname=app sslmode=require" \
{% endif %}{% if cookiecutter.include_workers == "y" %} \
    -e APP_REDIS_ADDR=redis:6379 \
{% endif %}    {{ cookiecutter.project_slug }}:latest

{% endif %}{% if cookiecutter.include_workers == "y" %}# Worker only
docker run -d \
    -e APP_LOG_LEVEL=info \
    -e APP_REDIS_ADDR=redis:6379 \
{% if cookiecutter.include_orm == "y" %} \
    -e APP_DATABASE_DRIVER=postgres \
    -e APP_DATABASE_DSN="host=db user=postgres password=SECRET dbname=app sslmode=require" \
{% endif %}    {{ cookiecutter.project_slug }}-worker:latest

{% endif %}{% if cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}# CLI (one-off, no -d flag)
docker run --rm \
    -e APP_DATABASE_DRIVER=postgres \
    -e APP_DATABASE_DSN="host=db user=postgres password=SECRET dbname=app sslmode=require" \
    {{ cookiecutter.project_slug }}-cli:latest \
    migrate up
{% endif %}
```

### Health Check

{% if cookiecutter.include_web == "y" %}The `/health` endpoint returns `{"status":"ok"}` for orchestration tools:

```bash
curl http://localhost:8080/health
```

Use this for Kubernetes liveness and readiness probes.{% endif %}

### Resource Limits

The production compose file includes resource limits by default:

| Service | Memory Limit | CPU Limit | Memory Reservation |
|---------|-------------|-----------|-------------------|
{% if cookiecutter.include_web == "y" %}| `web` | 512M | 1.0 | 128M |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `worker` | 256M | 0.5 | 64M |
{% endif %}

Go's GC and connection pool settings should be tuned to match these limits.

## Multi-Architecture Builds

The Docker workflow in CI builds for multiple architectures:

```yaml
platforms: linux/amd64,linux/arm64
```

This produces images for both x86_64 and ARM64 systems. The build uses QEMU emulation in CI to cross-compile.

### Local Multi-Arch Build

Build for a specific architecture locally:

```bash
docker buildx build --platform linux/arm64 --target web-build -t {{ cookiecutter.project_slug }}:arm64 .
```

## Docker Build Targets

| Just Command | Description |
|--------------|-------------|
| `docker-build` | Build the Docker image (all targets) |
| `docker-build-web` | Build only the web server image |
{% if cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}| `docker-build-cli` | Build only the CLI image |
{% endif %}{% if cookiecutter.include_workers == "y" %}| `docker-build-worker` | Build only the worker image |
{% endif %}| `docker-up` | Start development services (hot-reload) |
| `docker-down` | Stop all docker compose services |
{% if cookiecutter.include_web == "y" or cookiecutter.include_workers == "y" %}| `docker-up-prod` | Start production services (distroless) |
{% endif %}

```bash
just docker-build
just docker-up
just docker-down
```
