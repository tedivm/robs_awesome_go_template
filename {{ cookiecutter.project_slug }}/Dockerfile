# ---- Builder (shared) ----
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

# ---- Worker target ----
{% if cookiecutter.include_workers == "y" %}
FROM builder AS worker-build

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X {{ cookiecutter.import_path }}/internal/version.Version=${VERSION} -X {{ cookiecutter.import_path }}/internal/version.Commit=${COMMIT} -X {{ cookiecutter.import_path }}/internal/version.Date=${DATE}" \
    -o /app/worker \
    ./cmd/worker

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=worker-build /app/worker .

ENTRYPOINT ["./worker"]

# ---- CLI target ----
{% endif %}{% if cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}
FROM builder AS cli-build

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X {{ cookiecutter.import_path }}/internal/version.Version=${VERSION} -X {{ cookiecutter.import_path }}/internal/version.Commit=${COMMIT} -X {{ cookiecutter.import_path }}/internal/version.Date=${DATE}" \
    -o /app/cli \
    ./cmd/cli

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=cli-build /app/cli .

ENTRYPOINT ["./cli"]

# ---- Web target ----
{% endif %}{% if cookiecutter.include_web == "y" %}
FROM builder AS web-build

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X {{ cookiecutter.import_path }}/internal/version.Version=${VERSION} -X {{ cookiecutter.import_path }}/internal/version.Commit=${COMMIT} -X {{ cookiecutter.import_path }}/internal/version.Date=${DATE}" \
    -o /app/{{ cookiecutter.project_slug }} \
    ./cmd/{{ cookiecutter.project_slug }}

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=web-build /app/{{ cookiecutter.project_slug }} .
COPY --from=builder /src/static ./static

EXPOSE 8080

ENTRYPOINT ["./{{ cookiecutter.project_slug }}"]

{% endif %}