module {{ cookiecutter.import_path }}

go {{ cookiecutter.go_version }}

require (
{% if cookiecutter.include_web == "y" %}
	github.com/danielgtaylor/huma/v2 v2.40.0
	github.com/go-chi/chi/v5 v5.2.1
{% endif %}
{% if cookiecutter.include_orm == "y" %}
	github.com/uptrace/bun v1.3.0
	github.com/uptrace/bun/dialect/pgdialect v1.3.0
	github.com/uptrace/bun/dialect/sqlitedialect v1.3.0
	github.com/uptrace/bun/driver/pgdriver v1.3.0
	modernc.org/sqlite v1.34.1
{% endif %}
{% if cookiecutter.include_cli == "y" and cookiecutter.include_orm == "y" %}
	github.com/spf13/cobra v1.9.1
{% endif %}
{% if cookiecutter.include_workers == "y" %}
	github.com/hibiken/asynq v0.23.0
{% endif %}
{% if cookiecutter.include_web == "y" %}
	github.com/prometheus/client_golang v1.22.0
{% endif %}
	github.com/google/uuid v1.6.0
	github.com/knadh/koanf/parsers/yaml v1.1.1
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/v2 v2.4.1
	github.com/joho/godotenv v1.5.1
{% if cookiecutter.include_web == "y" %}
	github.com/golang-jwt/jwt/v5 v5.2.2
{% endif %}
	github.com/stretchr/testify v1.10.0
)
