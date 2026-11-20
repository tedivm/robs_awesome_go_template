---
name: managing-migrations
description: "Use when creating, applying, or rolling back database schema migrations with goose. Don't use for model creation or ORM queries."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+, goose v3"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Create migrations with `just migrate-create <name>`. Each migration is a single file with both `up` (apply) and `down` (rollback) sections. Apply with `just migrate-apply`.

## Workflow

1. Create migration: `just migrate-create <name>`
2. Write SQL with `-- +goose Up` and `-- +goose Down` annotations in the single file
3. Wrap multi-statement migrations in `-- +goose StatementBegin` / `-- +goose StatementEnd`
4. Test against both SQLite (dev) and PostgreSQL (prod)

### Creating a migration

```bash
# Create a new migration file (creates in both pg/ and sqlite/)
just migrate-create add_items

# Creates:
# migrations/pg/002_add_items.sql
# migrations/sqlite/002_add_items.sql
```

### Migration file format

**Single file with both directions:**

```sql
-- migrations/pg/002_add_items.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_items_status ON items(status);
CREATE INDEX idx_items_created_at ON items(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_created_at;
DROP INDEX IF EXISTS idx_items_status;
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
```

**Rules:**
- `-- +goose Up` and `-- +goose Down` annotations are required
- Wrap multi-statement migrations in `-- +goose StatementBegin` / `-- +goose StatementEnd`
- Down migrations must be the inverse of up migrations
- Use `IF EXISTS` / `IF NOT EXISTS` where possible for idempotency

### Running migrations

```bash
# Apply all pending migrations
just migrate-apply

# Check migration status
just migrate-plan

# Rollback the last migration
go run ./cmd/cli migrate down
```

### Environment variable substitution

goose supports env var substitution in SQL files using `${VAR_NAME}` syntax:

```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "${GOOSE_EXTENSION}";
```

Set the variable before running: `GOOSE_EXTENSION=uuid-ossp goose -dir migrations up`.

### Keeping migrations in sync with bun models

1. **Add a new model** in `internal/model/` with bun tags
2. **Create a migration** with `goose create` that matches the model's columns
3. **Verify the migration** runs against both SQLite (dev) and PostgreSQL (prod)
4. **Update the model** if the migration changes — they should always match

**Cross-database tips:**
- Use `UUID` type in PostgreSQL migrations; bun handles SQLite conversion
- Use `TIMESTAMPTZ` for timestamps; bun translates to SQLite's `TEXT` representation
- Use `gen_random_uuid()` in PostgreSQL; bun's `BeforeInsert` hook generates UUIDs for SQLite

## Key conventions

- One migration per logical change — don't combine unrelated changes
- Always write down migrations — every up must have a working down
- Never modify an applied migration — create a new one instead
- Name migrations descriptively: `add_items`, `add_user_roles`, `rename_email_column`
- Test migrations against both SQLite and PostgreSQL before committing
- Migrations live in `migrations/` — numbered sequentially from `001`
