#!/usr/bin/env bash
set -euo pipefail

{% if cookiecutter.include_orm == "y" %}
# Wait for database
until pg_isready -h db -p 5432 -U postgres; do
  sleep 1
done

{% endif %}{% if cookiecutter.include_orm == "y" and cookiecutter.include_cli == "y" %}
# Run migrations via CLI
go run ./cmd/cli migrate up

{% endif %}# Run the application
{% if cookiecutter.include_web == "y" %}
go run ./cmd/{{ cookiecutter.project_slug }}
{% elif cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}
go run ./cmd/cli
{% endif %}